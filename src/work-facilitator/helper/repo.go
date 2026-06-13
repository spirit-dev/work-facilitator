package helper

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
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
	separatorParam           = "separator"
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
	if gitRepoNs[0:1] == "/" {
		gitRepoNs = gitRepoNs[1:]
	}
	log.Debugln("gitRepoNs: " + gitRepoNs)
	gitRepoName := strings.Replace(nameS[len(nameS)-1], ".git", "", -1)
	log.Debugf("gitRepoName: %v\n", gitRepoName)
	gitRepoFName := strings.Join([]string{gitRepoNs, gitRepoName}, "/")
	gitRepoHost, _, _ := net.SplitHostPort(repoParsedUrl.Host)
	if gitRepoHost == "" {
		gitRepoHost = repoParsedUrl.Host
	}
	log.Debugf("gitRepoHost: %v\n", gitRepoHost)

	// Build browser url based on origin url
	browserUrl := "https://" + gitRepoHost + "/" + gitRepoNs + "/" + gitRepoName
	log.Debugln("Browser url: " + browserUrl)

	// Repo default branch
	defaultBranch, confInit, remoteGot := repoDefautltBranch(wfConfig.DefaultBranch, publicAuthKey)

	// Repo defined separator
	separator, err := repoConfigGetParam(wfsetupSection, separatorParam)
	if err != nil {
		separator = c.NOTGIVEN
	}

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
		Separator:           separator,
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

// RepoCheckUncommittedFiles checks for uncommitted/unstaged files in the repository.
// Returns a map of file paths to their status and a boolean indicating if any uncommitted files exist.
// This function detects:
//   - Unstaged changes (modified but not staged)
//   - Staged changes (staged but not committed)
//   - Untracked files
//
// The function respects .gitignore patterns (both repository and global).
func RepoCheckUncommittedFiles() (map[string]*git.FileStatus, bool, error) {
	w, err := repo.Worktree()
	if err != nil {
		return nil, false, err
	}

	status, err := w.Status()
	if err != nil {
		return nil, false, err
	}

	uncommittedFiles := make(map[string]*git.FileStatus)
	hasUncommitted := false

	for filePath, fileStatus := range status {
		// Check if file has any uncommitted changes
		// Staging != Unmodified means there are staged changes
		// Worktree != Unmodified means there are unstaged changes or untracked files
		if fileStatus.Staging != git.Unmodified || fileStatus.Worktree != git.Unmodified {
			// For untracked files, check if they should be ignored
			// Note: go-git's Status() already filters out ignored files for untracked files
			// but we double-check here for safety
			if fileStatus.Worktree == git.Untracked {
				// Status() in go-git already respects .gitignore for untracked files
				// If a file appears in status as Untracked, it's not ignored
				log.Debugf("Untracked file detected: %s\n", filePath)
			}

			uncommittedFiles[filePath] = fileStatus
			hasUncommitted = true
			log.Debugf("Uncommitted file detected: %s (Staging: %c, Worktree: %c)\n",
				filePath, fileStatus.Staging, fileStatus.Worktree)
		}
	}

	return uncommittedFiles, hasUncommitted, nil
}

func RepoAddAllFiles(ignorePatterns []*regexp.Regexp) {
	log.Debugln("git add .")

	w, err := repo.Worktree()
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}

	// Get current status to identify files to add
	status, err := w.Status()
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}

	// Collect all modified/untracked files and deleted files
	var filesToAdd []string
	var filesToRemove []string

	for filePath, fileStatus := range status {
		// Handle deleted files separately
		if fileStatus.Worktree == git.Deleted {
			filesToRemove = append(filesToRemove, filePath)
		} else if fileStatus.Worktree == git.Modified || fileStatus.Worktree == git.Added || fileStatus.Worktree == git.Untracked {
			// Add files that are modified, added, or untracked
			filesToAdd = append(filesToAdd, filePath)
		}
	}

	// Filter out ignored files from additions
	filesToAdd = FilterIgnoredFiles(filesToAdd, ignorePatterns)
	// Filter out ignored files from deletions
	filesToRemove = FilterIgnoredFiles(filesToRemove, ignorePatterns)

	// Add each file individually
	for _, filePath := range filesToAdd {
		log.Debugf("Adding file: %s\n", filePath)
		_, err := w.Add(filePath)
		if err != nil {
			log.Warningf("Failed to add file '%s': %v\n", filePath, err)
		}
	}

	// Remove (stage deletion of) each deleted file
	for _, filePath := range filesToRemove {
		log.Debugf("Staging deletion of file: %s\n", filePath)
		_, err := w.Remove(filePath)
		if err != nil {
			log.Warningf("Failed to stage deletion of file '%s': %v\n", filePath, err)
		}
	}
}

func RepoCommit(message string, ignorePatterns []*regexp.Regexp) {
	log.Debugln("git commit -m " + message)

	w, err := repo.Worktree()
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}

	// Check for staged files matching ignore patterns
	// This is a safety check in case files were manually staged with `git add`
	// (RepoAddAllFiles already filters these out, but this catches manual staging)
	status, err := w.Status()
	if err != nil {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Fatalln(err)
	}

	stagedIgnored := GetStagedIgnoredFiles(status, ignorePatterns)
	if len(stagedIgnored) > 0 {
		if !Quiet {
			SpinStopDisplay("fail")
		}
		log.Warningln("Found staged files matching ignore patterns:")
		for _, f := range stagedIgnored {
			log.Warningln("  - " + f)
		}
		log.Warningln("")
		log.Warningln("Please unstage these files manually using:")
		log.Warningln("  git reset HEAD <file>")
		log.Fatalln("Commit aborted: ignored files are staged")
	}

	// Proceed with commit
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

// GetStagedDiff returns the git diff of staged changes (HEAD vs index blobs).
// It reads staged content from git index blobs, not the working tree,
// ensuring unstaged modifications never leak into the diff.
func GetStagedDiff() (string, error) {
	return getDiff(false) // false = read from index blobs
}

// GetWorkingTreeDiff returns the git diff including working-tree modifications
// for staged files (HEAD vs working directory). Used with the -U flag.
// Only files that have staged entries in the index are included;
// files not in the index are excluded regardless of modifications.
func GetWorkingTreeDiff() (string, error) {
	return getDiff(true) // true = read from working tree
}

// getDiff is the core diff function. When useWorktree is false, it reads staged
// file content from git index blobs. When true, it reads from the working tree.
func getDiff(useWorktree bool) (string, error) {
	w, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	// Check if there are any staged changes
	hasStagedChanges := false
	for _, fileStatus := range status {
		if fileStatus.Staging != git.Unmodified && fileStatus.Staging != git.Untracked {
			hasStagedChanges = true
			break
		}
	}

	if !hasStagedChanges {
		return "", fmt.Errorf("no staged changes found")
	}

	// Get HEAD commit
	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	headCommit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	headTree, err := headCommit.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD tree: %w", err)
	}

	// Get index (staged changes)
	idx, err := repo.Storer.Index()
	if err != nil {
		return "", fmt.Errorf("failed to get index: %w", err)
	}

	var diffBuilder strings.Builder

	for _, entry := range idx.Entries {
		filePath := entry.Name

		// Get file from HEAD
		headFile, err := headTree.File(filePath)
		var headContent string
		if err == nil {
			headContent, _ = headFile.Contents()
		}

		// Get staged/new content: from index blob or working tree
		var newContent string
		if useWorktree {
			newContent = readWorkingTreeFile(w, filePath)
		} else {
			newContent = readIndexBlob(filePath, entry.Hash)
		}

		// Only include if there's a difference
		if headContent != newContent {
			diffBuilder.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", filePath, filePath))
			diffBuilder.WriteString(fmt.Sprintf("--- a/%s\n", filePath))
			diffBuilder.WriteString(fmt.Sprintf("+++ b/%s\n", filePath))

			headLines := strings.Split(headContent, "\n")
			newLines := strings.Split(newContent, "\n")

			diffBuilder.WriteString(formatUnifiedDiff(headLines, newLines, 3))
		}
	}

	diff := diffBuilder.String()
	if diff == "" {
		return "", fmt.Errorf("no diff generated")
	}

	log.Debugln("Generated diff, length:", len(diff))
	return diff, nil
}

// readIndexBlob reads a staged file's content from the git object store (index blob).
func readIndexBlob(filePath string, hash plumbing.Hash) string {
	obj, err := repo.Storer.EncodedObject(plumbing.BlobObject, hash)
	if err != nil {
		log.Debugf("Failed to get blob for %s: %v\n", filePath, err)
		return ""
	}

	reader, err := obj.Reader()
	if err != nil {
		log.Debugf("Failed to read blob for %s: %v\n", filePath, err)
		return ""
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		log.Debugf("Failed to read blob content for %s: %v\n", filePath, err)
		return ""
	}

	return string(content)
}

// readWorkingTreeFile reads a file's current content from the working tree.
func readWorkingTreeFile(w *git.Worktree, filePath string) string {
	f, err := w.Filesystem.Open(filePath)
	if err != nil {
		log.Debugf("Failed to open file %s: %v\n", filePath, err)
		return ""
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		log.Debugf("Failed to read file %s: %v\n", filePath, err)
		return ""
	}

	return string(content)
}

// diffOp represents a single diff operation
type diffOp struct {
	action byte   // '+', '-', or ' '
	line   string
}

// formatUnifiedDiff produces a unified diff string with context lines.
// contextLines specifies how many unchanged lines to show around each hunk.
func formatUnifiedDiff(oldLines, newLines []string, contextLines int) string {
	if len(oldLines) == 0 && len(newLines) == 0 {
		return ""
	}

	ops := computeDiffOps(oldLines, newLines)
	hunks := buildHunks(ops, contextLines)

	var buf strings.Builder
	for _, h := range hunks {
		// Hunk header: @@ -oldStart,oldCount +newStart,newCount @@
		buf.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
			h.oldStart+1, h.oldCount,
			h.newStart+1, h.newCount))

		for _, line := range h.lines {
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}

	return buf.String()
}

// hunk represents a unified diff hunk
type hunk struct {
	oldStart, oldCount int
	newStart, newCount int
	lines              []string // prefixed with ' ', '-', '+'
}

// region represents a range of indices in the diff ops
type region struct{ start, end int }

// buildHunks groups diff operations into hunks with context
func buildHunks(ops []diffOp, contextLines int) []hunk {
	if len(ops) == 0 {
		return nil
	}

	// Find change regions (sequences with at least one '+' or '-')
	var regions []region

	for i := 0; i < len(ops); i++ {
		if ops[i].action != ' ' {
			// Found a change
			start := i
			for i < len(ops) && (ops[i].action != ' ' ||
				(i+1 < len(ops) && ops[i+1].action != ' ' && i-start < contextLines*2)) {
				i++
			}
			regions = append(regions, region{start, i})
		}
	}

	// Merge overlapping/nearby regions
	merged := mergeRegions(regions, contextLines)

	// Build hunks
	var hunks []hunk
	for _, r := range merged {
		// Expand to include context
		start := r.start - contextLines
		if start < 0 {
			start = 0
		}
		end := r.end + contextLines
		if end > len(ops) {
			end = len(ops)
		}

		// Compute hunk metrics
		var oldStart, oldCount, newStart, newCount int
		oldStartSet, newStartSet := false, false
		var lines []string

		for i := start; i < end; i++ {
			op := ops[i]
			switch op.action {
			case ' ':
				if !oldStartSet {
					oldStart = findOldPrefixCount(ops, start)
					oldStartSet = true
				}
				if !newStartSet {
					newStart = findNewPrefixCount(ops, start)
					newStartSet = true
				}
				oldCount++
				newCount++
				lines = append(lines, " "+op.line)
			case '-':
				if !oldStartSet {
					oldStart = findOldPrefixCount(ops, start)
					oldStartSet = true
				}
				if !newStartSet {
					newStart = findNewPrefixCount(ops, start)
					newStartSet = true
				}
				oldCount++
				lines = append(lines, "-"+op.line)
			case '+':
				if !oldStartSet {
					oldStart = findOldPrefixCount(ops, start)
					oldStartSet = true
				}
				if !newStartSet {
					newStart = findNewPrefixCount(ops, start)
					newStartSet = true
				}
				newCount++
				lines = append(lines, "+"+op.line)
			}
		}

		hunks = append(hunks, hunk{
			oldStart: oldStart,
			oldCount: oldCount,
			newStart: newStart,
			newCount: newCount,
			lines:    lines,
		})
	}

	return hunks
}

// findOldPrefixCount counts lines in old that appear before the given index
func findOldPrefixCount(ops []diffOp, endIdx int) int {
	count := 0
	for i := 0; i < endIdx && i < len(ops); i++ {
		if ops[i].action == ' ' || ops[i].action == '-' {
			count++
		}
	}
	return count
}

// findNewPrefixCount counts lines in new that appear before the given index
func findNewPrefixCount(ops []diffOp, endIdx int) int {
	count := 0
	for i := 0; i < endIdx && i < len(ops); i++ {
		if ops[i].action == ' ' || ops[i].action == '+' {
			count++
		}
	}
	return count
}

// mergeRegions merges regions that overlap or are close enough to share context
func mergeRegions(regions []region, contextLines int) []region {
	if len(regions) == 0 {
		return nil
	}

	merged := []region{regions[0]}
	for i := 1; i < len(regions); i++ {
		last := &merged[len(merged)-1]
		if regions[i].start-contextLines <= last.end+contextLines {
			// Overlap or close enough: merge
			if regions[i].end > last.end {
				last.end = regions[i].end
			}
		} else {
			merged = append(merged, regions[i])
		}
	}
	return merged
}

// computeDiffOps computes diff operations between old and new lines
func computeDiffOps(oldLines, newLines []string) []diffOp {
	table := buildLCSTable(oldLines, newLines)
	ops := backtrackOps(table, oldLines, newLines, len(oldLines), len(newLines))
	return ops
}

// buildLCSTable builds the LCS dynamic programming table
func buildLCSTable(oldLines, newLines []string) [][]int {
	m, n := len(oldLines), len(newLines)
	table := make([][]int, m+1)
	for i := range table {
		table[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if oldLines[i-1] == newLines[j-1] {
				table[i][j] = table[i-1][j-1] + 1
			} else if table[i-1][j] >= table[i][j-1] {
				table[i][j] = table[i-1][j]
			} else {
				table[i][j] = table[i][j-1]
			}
		}
	}

	return table
}

// backtrackOps reconstructs the diff operations from the LCS table
func backtrackOps(table [][]int, oldLines, newLines []string, i, j int) []diffOp {
	if i == 0 && j == 0 {
		return nil
	}

	if i > 0 && j > 0 && oldLines[i-1] == newLines[j-1] {
		ops := backtrackOps(table, oldLines, newLines, i-1, j-1)
		ops = append(ops, diffOp{' ', oldLines[i-1]})
		return ops
	}

	if j > 0 && (i == 0 || table[i][j-1] >= table[i-1][j]) {
		ops := backtrackOps(table, oldLines, newLines, i, j-1)
		ops = append(ops, diffOp{'+', newLines[j-1]})
		return ops
	}

	ops := backtrackOps(table, oldLines, newLines, i-1, j)
	ops = append(ops, diffOp{'-', oldLines[i-1]})
	return ops
}

// FilterDiffByPatterns filters out files matching exclude patterns from the diff
func FilterDiffByPatterns(diff string, excludePatterns []string) string {
	if len(excludePatterns) == 0 {
		return diff
	}

	// Compile patterns
	var patterns []*regexp.Regexp
	for _, pattern := range excludePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			log.Warningln("Invalid exclude pattern:", pattern)
			continue
		}
		patterns = append(patterns, re)
	}

	if len(patterns) == 0 {
		return diff
	}

	// Split diff by file
	lines := strings.Split(diff, "\n")
	var filteredLines []string
	var currentFile string
	var skipCurrentFile bool

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			// Extract filename
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				currentFile = strings.TrimPrefix(parts[2], "a/")

				// Check if file should be excluded
				skipCurrentFile = false
				for _, pattern := range patterns {
					if pattern.MatchString(currentFile) {
						skipCurrentFile = true
						log.Debugln("Excluding file from diff:", currentFile)
						break
					}
				}
			}
		}

		if !skipCurrentFile {
			filteredLines = append(filteredLines, line)
		}
	}

	return strings.Join(filteredLines, "\n")
}
