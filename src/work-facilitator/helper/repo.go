package helper

import (
	"errors"
	"fmt"
	"os"
	c "spirit-dev/work-facilitator/work-facilitator/common"
	"strconv"
	"strings"
	"time"

	giturls "github.com/whilp/git-urls"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	log "github.com/sirupsen/logrus"
)

var (
	repo    *git.Repository
	repoCfg *config.Config
	err     error

	Quiet = false
)

const (
	wfsetupSection   = "workflowsetup"
	wfsetupSectionV2 = "workflow.setup"
	wfSection        = "workflow"
	branchSection    = "branch"

	defaultBranchParam       = "default-branch"
	defaultBranchExpiryParam = "default-branch-expiry"
	currentParam             = "current"
	typeBranchParam          = "type-branch"
	typeCommitParam          = "type-commit"
	mrRefParam               = "mrref"
	ticketParam              = "ticket"
	titleParam               = "title"
	branchParm               = "branch"
	commitParam              = "commit"
	REFBRANCHPARAM           = "refbranch"
	remoteParam              = "remote"
	mergeParam               = "merge"
	vscodeMergeBaseParam     = "vscode-merge-base"

	expiryLayout    = "2006-01-02 15:04:05"
	expiryUtcLayout = expiryLayout + " +0000 UTC"

	originValue = "origin"
)

// Public functions

func NewRepo(wfConfig c.Config) c.Repo {

	if !Quiet {
		SpinStartDisplay("Repository config")
	}

	// Open repo
	openRepo()

	// Build auth is ssh key is set
	var publicAuthKey *ssh.PublicKeys
	if wfConfig.HasSshKeyId {
		var password string
		publicAuthKey, err = ssh.NewPublicKeysFromFile("git", wfConfig.SshKeyId, password)
		if err != nil {
			if !Quiet {
				SpinStopDisplay("fail")
			}
			log.Fatalln(err)
		}
	}

	// repo base path
	repoBasePath := repoBasePath()

	// Verify config
	testRepo(repoBasePath)

	// Origin URL
	originUrl := repoOriginUrl()
	// Browser URL
	// Parse Git URL
	repoParsedUrl, err := giturls.Parse(originUrl)
	if err != nil {

		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}
	nameS := strings.Split(repoParsedUrl.Path, "/")

	// Extract git repo name
	gitRepoNs := strings.Join(remove(nameS, len(nameS)-1), "/")
	log.Debugln("Repo NS: " + gitRepoNs)
	gitRepoName := strings.Replace(nameS[len(nameS)-1], ".git", "", -1)
	log.Debugf("gitRepoName: %v\n", gitRepoName)
	gitRepoFName := strings.Join([]string{gitRepoNs, gitRepoName}, "/")

	// Build browser url based on origin url
	browserUrl := "https://" + repoParsedUrl.Host + "/" + gitRepoNs + "/" + gitRepoName
	log.Debugln("Browser url: " + browserUrl)

	// Repo default branch
	defaultBranch, confInit, remoteGot := repoDefautltBranch(wfConfig.DefaultBranch, publicAuthKey)

	// Generate work list
	worklist := repoConfigGenerateWorklist()

	// Some diplays nicely

	if !Quiet {
		SpinStopDisplay("success")
		if confInit {
			SpinSideNoteDisplay("initialized .git/config file")
		}
		if remoteGot {
			SpinSideNoteDisplay("updated remote branch")
		}
	}

	// Current workflow
	currentWf, err := repoConfigGetParam(wfsetupSection, currentParam)
	hasCurrentWf := true
	if err != nil || currentWf == "" {
		hasCurrentWf = false
	}

	var currentWfData c.Workflow
	if hasCurrentWf {
		currentWfData = RepoConfigGetCurrentWorkflow(currentWf)
	}

	repoInfo := c.Repo{
		BasePath:            repoBasePath,
		OriginUrl:           originUrl,
		BrowserUrl:          browserUrl,
		Namespace:           gitRepoNs,
		Name:                gitRepoName,
		FName:               gitRepoFName,
		DefaultBranch:       defaultBranch,
		Worklist:            worklist,
		CurrentWorkflowName: currentWf,
		HasCurrentWorkflow:  hasCurrentWf,
		CurrentWorkflowData: currentWfData,
		PublicAuthKey:       publicAuthKey,
	}

	return repoInfo
}

func RepoConfigGetCurrentWorkflow(currentWf string) c.Workflow {
	issue, _ := strconv.Atoi(repoGetWorkflowParam(currentWf, mrRefParam))
	return c.Workflow{
		CurrentWork: currentWf,
		BranchType:  repoGetWorkflowParam(currentWf, typeBranchParam),
		CommitType:  repoGetWorkflowParam(currentWf, typeCommitParam),
		Issue:       issue,
		Ticket:      repoGetWorkflowParam(currentWf, ticketParam),
		Title:       repoGetWorkflowParam(currentWf, titleParam),
		Commit:      repoGetWorkflowParam(currentWf, commitParam),
		RefBranch:   repoGetWorkflowParam(currentWf, REFBRANCHPARAM),
		Branch:      repoGetWorkflowParam(currentWf, branchParm),
	}
}

func RepoConfigDefineCurrentWorkflow(workflow string) {
	repoConfigUpdateParam(wfsetupSection, currentParam, workflow)
}

func RepoConfigDefineWorkflow(rootCfg c.Config, wf c.Workflow) {
	// Set the current work flow
	RepoConfigDefineCurrentWorkflow(wf.CurrentWork)

	repoConfigAddSubSectParam(wfSection, wf.CurrentWork, typeBranchParam, wf.BranchType)
	repoConfigAddSubSectParam(wfSection, wf.CurrentWork, typeCommitParam, wf.CommitType)
	repoConfigAddSubSectParam(wfSection, wf.CurrentWork, titleParam, wf.Title)
	repoConfigAddSubSectParam(wfSection, wf.CurrentWork, branchParm, wf.CurrentWork)
	repoConfigAddSubSectParam(wfSection, wf.CurrentWork, commitParam, wf.Commit)
	repoConfigAddSubSectParam(wfSection, wf.CurrentWork, REFBRANCHPARAM, wf.RefBranch)

	// Set workflow variables
	if rootCfg.Ticketing == c.GITLAB {
		repoConfigAddSubSectParam(wfSection, wf.CurrentWork, mrRefParam, strconv.Itoa(wf.Issue))

	}
	if rootCfg.Ticketing == c.JIRA {
		repoConfigAddSubSectParam(wfSection, wf.CurrentWork, ticketParam, wf.Ticket)
	}

	// Set Branch variables
	repoCfg.Branches[wf.CurrentWork] = &config.Branch{
		Name:   wf.CurrentWork,
		Remote: originValue,
		Merge:  plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", wf.CurrentWork)),
	}

	repoConfigAddSubSectParam(branchSection, wf.CurrentWork, vscodeMergeBaseParam, fmt.Sprintf("origin/%s", wf.CurrentWork)) // Not sure this one works
}

func RepoConfigWrite() {
	err := repo.SetConfig(repoCfg)
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}
}

func WorkflowExisting(branch string) bool {
	// Negate the returned value if workflow exist or not
	return repoCfg.Raw.Section(wfSection).HasSubsection(branch)
}

func RepoCheckout(branch string, pubKey *ssh.PublicKeys) {

	w, err := repo.Worktree()
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}

	// Check if a branch exists
	branchExists := branchExists(branch)

	// ... checking out branch
	log.Debug("git checkout " + branch)

	// Create new Reference
	branchRefName := plumbing.NewBranchReferenceName(branch)
	// Set checkout Options
	branchCoOpts := git.CheckoutOptions{
		Branch: plumbing.ReferenceName(branchRefName),
		Force:  false,
		Keep:   true,
		Create: !branchExists, // Equivalent of git checkout -b if branch does not exists locally
	}
	if err := w.Checkout(&branchCoOpts); err != nil {
		log.Warningf("local checkout of branch '%s' failed, will attempt to fetch remote branch of same name.\n", branch)
		log.Warningln("like `git checkout <branch>` defaulting to `git checkout -b <branch> --track <remote>/<branch>`")

		mirrorRemoteBranchRefSpec := fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch)
		err = fetchOrigin(mirrorRemoteBranchRefSpec, pubKey)
		if err != nil {
			if !Quiet {
				SpinStopDisplay("fail")
			}
			log.Fatalln(err)
		}

		err = w.Checkout(&branchCoOpts)
		if err != nil {
			if !Quiet {
				SpinStopDisplay("fail")
			}
			log.Fatalln(err)
		}
	}
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}
}

func RepoPull(pubKey *ssh.PublicKeys) string {
	// Get the working directory for the repository
	w, err := repo.Worktree()
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}

	opts := &git.PullOptions{
		RemoteName: originValue,
		Force:      true,
	}
	if pubKey != nil {
		opts.Auth = pubKey
	}

	var info string
	// Pull the latest changes from the origin remote and merge into the current branch
	log.Debugln("git pull origin")
	err = w.Pull(opts)
	if err != nil {
		log.Debugln(err)
		info = err.Error()
	}

	// Print the latest commit that was just pulled
	ref, err := repo.Head()
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}

	log.Debugln(commit)
	return info
}

func RepoGetWorkflowParam(subsection, param string) (string, error) {
	return repoConfigGetSubParam(wfSection, subsection, param)
}

func RepoDeleteBranch(branch string, ref plumbing.ReferenceName) {
	// Ensure branch exists
	_, err := repo.Branch(branch)
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln("Branch " + branch + " does not exists")
	}

	// Delete branch
	log.Debugln("git branch -D " + branch)
	repo.DeleteBranch(branch)
	err = repo.Storer.RemoveReference(ref)
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}
}

func RepoHead() *plumbing.Reference {
	h, err := repo.Head()
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}

	return h
}

func RepoGetBranchRef(branch string) plumbing.ReferenceName {
	ref := plumbing.NewBranchReferenceName(branch)
	log.Debugf("ref: %v\n", ref)
	return ref
}

func RepoConfigDeleteWorkflow(workflow string) {
	repoCfg.Raw.Section(wfSection).RemoveSubsection(workflow)
}

func RepoConfigDeleteBranch(workflow string) {
	delete(repoCfg.Branches, workflow)
}

func RepoConfigDeleteCurrentWorkflow() {
	repoCfg.Raw.Section(wfsetupSection).RemoveOption(currentParam)
}

func RepoConfigRefresh() {
	repoCfg, err = repo.Config()
}

func RepoStatus() git.Status {
	w, err := repo.Worktree()
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}
	s, err := w.Status()
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}
	return s
}

func RepoAddAllFiles() {
	log.Debugln("git add .")

	w, err := repo.Worktree()
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}

	erro := w.AddWithOptions(&git.AddOptions{
		All: true,
	})
	if erro != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(erro)
	}
}

func RepoCommit(message string) {
	log.Debugln("git commit -m " + message)

	w, err := repo.Worktree()
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}
	_, err = w.Commit(message, &git.CommitOptions{
		AllowEmptyCommits: false,
	})
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}
}

func RepoPush(pubKey *ssh.PublicKeys, branch string) {
	log.Debugln("git push origin")

	refSpec := config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch))
	log.Debugf("refSpec.String(): %v\n", refSpec.String())
	opts := &git.PushOptions{
		RemoteName: originValue,
		RefSpecs:   []config.RefSpec{refSpec},
	}
	if pubKey != nil {
		opts.Auth = pubKey
	}

	err = repo.Push(opts)
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}
}

// Local functions
//
//

func remove(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}

func repoGetWorkflowParam(subsection, param string) string {
	p, _ := RepoGetWorkflowParam(subsection, param)
	return p
}

func repoConfigGenerateWorklist() []string {

	var workflow []string

	for _, s := range repoCfg.Raw.Section("workflow").Subsections {
		workflow = append(workflow, s.Name)
	}

	return workflow
}

func repoDefautltBranch(wfDefaultBranch string, pubKey *ssh.PublicKeys) (string, bool, bool) {

	confInit := false
	remoteGot := false
	computedDefaultBranch := wfDefaultBranch

	// Ensure the default workflowsetup section is set
	if !repoConfigHasParam(wfsetupSection, defaultBranchParam) {
		repoConfigInitWfSetup()
		confInit = true
	}

	if repoConfigHasParam(wfsetupSection, defaultBranchParam) {
		if repoDefaultBranchExpired() {
			// Get default branch from remote
			branch, erro := repoGetRemoteDefaultBranch(pubKey)
			// Fallback on workflow config to set the default branch
			if erro != nil {
				log.Warningln("Could not find default branch on remote")
				log.Warningln("Falling back on default branch set in .workflow.ini file")
				log.Warningln(erro.Error())
				computedDefaultBranch = wfDefaultBranch
				return computedDefaultBranch, confInit, true
			}
			// Update returned values
			remoteGot = true
			computedDefaultBranch = branch

			// If we go here, the remote repo delivered a default branch
			// Update local git config default-branch and default-branch-expiry
			repoConfigUpdateDefaultBranch(computedDefaultBranch)

		} else {
			computedDefaultBranch, _ = repoConfigGetParam(wfsetupSection, defaultBranchParam)
		}
	}
	return computedDefaultBranch, confInit, remoteGot
}

func repoConfigUpdateDefaultBranch(branch string) {

	dtNow := time.Now()
	toAdd := 24 * 30 * time.Hour
	nDtNow := dtNow.Add(toAdd).Format(expiryLayout)
	repoConfigUpdateParam(wfsetupSection, defaultBranchParam, branch)
	repoConfigUpdateParam(wfsetupSection, defaultBranchExpiryParam, nDtNow)

	RepoConfigWrite()
}

func repoConfigUpdateParam(section string, param string, value string) {
	// Add section if not present
	if !repoCfg.Raw.HasSection(section) {
		repoCfg.Raw.Sections = append(repoCfg.Raw.Sections, config.NewConfig().Raw.Section(section))
	}
	repoCfg.Raw.Section(section).SetOption(param, value)
}

func repoConfigAddSubSectParam(section, subsection, param, value string) {
	// Add section if not present
	if !repoCfg.Raw.HasSection(section) {
		repoCfg.Raw.Sections = append(repoCfg.Raw.Sections, config.NewConfig().Raw.Section(section))
	}

	// Add subsection if not present
	if repoCfg.Raw.HasSection(section) && !repoCfg.Raw.Section(section).HasSubsection(subsection) {
		repoCfg.Raw.Section(section).Subsections = append(repoCfg.Raw.Section(section).Subsections, config.NewConfig().Raw.Section(section).Subsection(subsection))
	}

	repoCfg.Raw.Section(section).Subsection(subsection).AddOption(param, value)
}

func repoConfigInitWfSetup() error {

	// Add section
	// Add default option with empty values as such
	// ; .git/config
	// [workflowsetup]
	//   default-branch =
	//   default-branch-expiry = 2006-01-02 15:04:05
	//   current =
	repoCfg.Raw.Sections = append(repoCfg.Raw.Sections, config.NewConfig().Raw.Section(wfsetupSection))
	repoCfg.Raw.Section(wfsetupSection).AddOption(defaultBranchParam, "")
	repoCfg.Raw.Section(wfsetupSection).AddOption(defaultBranchExpiryParam, expiryLayout)

	// Write local config
	RepoConfigWrite()

	return err
}

func repoGetRemoteDefaultBranch(pubKey *ssh.PublicKeys) (string, error) {

	rem, remErr := repo.Remote(originValue)
	if remErr != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		return "", remErr
	}

	opts := &git.ListOptions{
		PeelingOption: git.IgnorePeeled,
	}
	if pubKey != nil {
		opts.Auth = pubKey
	}

	// We can then use every Remote functions to retrieve wanted information
	refs, remErr := rem.List(opts)
	if remErr != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		return "", remErr
	}
	for _, r := range refs {
		if r.Hash().IsZero() {
			defBranchSplt := strings.Split(r.Target().String(), "/")
			defBranch := defBranchSplt[len(defBranchSplt)-1]
			log.Debugln("Remotely got default branch: " + defBranch)
			return defBranch, nil
		}
	}
	return "", errors.New("no remote head found")
}

func repoDefaultBranchExpired() bool {
	if repoConfigHasParam(wfsetupSection, defaultBranchExpiryParam) {

		dtConfigParam, _ := repoConfigGetParam(wfsetupSection, defaultBranchExpiryParam)
		dtConfigStr, dtErr := time.Parse(expiryLayout, dtConfigParam)
		if dtErr != nil {
			if !Quiet {
				SpinStopDisplay("fail")
			}
			log.Fatalln(dtErr.Error())
		}
		dtConfig := dtConfigStr.Format(expiryUtcLayout)

		dtNow := time.Now().Format(expiryUtcLayout)

		log.Traceln("dtConfig: " + dtConfig + " < dtNow : " + dtNow)
		expired := dtConfig < dtNow
		if expired {
			log.Debugln("Git config default branch: expired")
		} else {
			log.Debugln("Git config default branch: not expired")
		}

		return expired
	}
	return false
}

func repoConfigGetParam(section string, option string) (string, error) {
	if repoConfigHasParam(section, option) {
		return repoCfg.Raw.Section(section).Option(option), nil
	}
	return "", errors.New("Non existing section '" + section + "' or option '" + option + "'")
}

func repoConfigGetSubParam(section, subsection, option string) (string, error) {

	if repoCfg.Raw.HasSection(section) {
		if repoCfg.Raw.Section(section).HasSubsection(subsection) {
			if repoCfg.Raw.Section(section).Subsection(subsection).HasOption(option) {
				return repoCfg.Raw.Section(section).Subsection(subsection).Option(option), nil
			} else {
				return "", errors.New("Non existing option '" + option + "' in subsection '" + section + " " + subsection + "'")
			}
		} else {
			return "", errors.New("Non existing subsection '" + subsection + "' in section '" + section + "'")
		}
	} else {
		return "", errors.New("Non existing section '" + section)
	}
}

func repoConfigHasParam(section string, option string) bool {
	if repoCfg.Raw.HasSection(section) {
		if repoCfg.Raw.Section(section).HasOption(option) {
			return true
		}
	}
	return false
}

func repoOriginUrl() string {
	// Get Origin url from git repo config
	oUrl := repoCfg.Remotes[originValue].URLs[0]
	log.Debugln("Origin url: " + oUrl)

	return oUrl
}

func testRepo(basePath string) {
	// test viability of the config (.git/config)
	repoCfg, err = repo.Config()
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Warningln("Error in the file " + basePath + "/.git/config:" + err.Error())
		log.Warningln("Attempt a .git/config upgrade from v2 to v3 ?")
		up := UpgradeV2ToV3()
		if up {
			upgradeV2toV3(basePath)
			log.Infoln("Upgrade done. Please retry.")
		}
		os.Exit(1)
	}
}

func upgradeV2toV3(basePath string) {
	gitCfgFile := basePath + "/.git/config"

	// Read file and replace workflow
	t, _ := os.ReadFile(gitCfgFile)
	t2 := strings.Replace(string(t), wfsetupSectionV2, wfsetupSection, -1)

	// Write file
	f, _ := os.OpenFile(gitCfgFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	f.WriteString(t2 + "\n")
}

func openRepo() {

	dir := CurrentPath()

	repo, err = git.PlainOpenWithOptions(dir, &git.PlainOpenOptions{
		DetectDotGit:          true,
		EnableDotGitCommonDir: true,
	})
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}
}

func repoBasePath() string {

	wt, _ := repo.Worktree()
	dir := wt.Filesystem.Root()
	log.Debugln("Repo found: " + dir)

	return dir
}

func fetchOrigin(refSpecStr string, pubKey *ssh.PublicKeys) error {
	remote, err := repo.Remote(originValue)
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}

	var refSpecs []config.RefSpec
	if refSpecStr != "" {
		refSpecs = []config.RefSpec{config.RefSpec(refSpecStr)}
	}

	opts := &git.FetchOptions{
		RefSpecs: refSpecs,
	}
	if pubKey != nil {
		opts.Auth = pubKey
	}

	if err = remote.Fetch(opts); err != nil {
		if err == git.NoErrAlreadyUpToDate {
			log.Debugln("refs already up to date")
		} else {
			return fmt.Errorf("fetch origin failed: %v", err)
		}
	}

	return nil
}

func branchExists(branch string) bool {

	branches, err := repo.Branches()
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}
	branchExists := false
	branches.ForEach(func(ref *plumbing.Reference) error {
		refName := strings.Split(ref.Name().String(), "refs/heads/")[1]
		if branch == refName {
			branchExists = true
		}
		return nil
	})

	log.Debugln(fmt.Sprintf("branchExists: %v", branchExists))

	return branchExists
}
