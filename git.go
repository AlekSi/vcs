package vcs

// import (
// 	"errors"
// 	"fmt"
// 	"os"
// 	"os/exec"
// 	"path/filepath"
// 	"regexp"
// 	"strings"
// )

// type GitVCS struct{}

// var _ VCS = &GitVCS{}

// var (
// 	WrongVCS            error = errors.New("Wrong VCS detected")
// 	CannotDetermineRepo error = errors.New("Unable to determine repository")
// )

// var remoteRegex = regexp.MustCompile("^origin\\s+(\\S+)\\s+\\S+$")

// // returns the currently checked out remote repository
// // according to the state of the working directory
// func (g *GitVCS) currentRepository(directory string) (string, error) {

// 	// Make sure to stop bubbling up the directory structures. If the
// 	// repo is another VCS sype, such as HG, git will try to bubble
// 	// up to a parent git repo.
// 	location, err := exec.Command("git", "rev-parse", "--git-dir").CombinedOutput()
// 	if err != nil {
// 		return "", WrongVCS
// 	}
// 	absPath, err := filepath.Abs(string(location))
// 	repoDir := filepath.Dir(absPath)
// 	if repoDir != directory {
// 		return "", WrongVCS
// 	}

// 	out, err := exec.Command("git", "remote", "-v").CombinedOutput()
// 	if err != nil {
// 		return "", WrongVCS
// 	}

// 	for _, line := range strings.Split(string(out), "\n") {
// 		if m := remoteRegex.FindStringSubmatch(line); len(m) == 2 {
// 			return m[1], nil
// 		}
// 	}

// 	return "", CannotDetermineRepo
// }

// func (g *GitVCS) currentCheckoutMatchesTagOrRef(ref string) bool {
// 	out, err := exec.Command("git", "log", "-n", "1", "--pretty=format:%H%d").CombinedOutput()
// 	if err != nil {
// 		return false
// 	}
// 	parts := strings.SplitN(string(out), " ", 2)

// 	sha := parts[0]
// 	if strings.HasPrefix(sha, ref) {
// 		return true
// 	}

// 	if len(parts) > 1 && strings.Contains(parts[1], "tag: ") {
// 		re := regexp.MustCompile("tag: ([^,)]*)")
// 		for _, tagMatch := range re.FindAllStringSubmatch(parts[1], 10) {
// 			if len(tagMatch) == 2 && tagMatch[1] == ref {
// 				return true
// 			}
// 		}
// 	}

// 	return false
// }

// // GitGet implements the getting logic for Git.
// func (g *GitVCS) Get(dep *Dependency) error {
// 	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)
// 	//Info("Cloning %s into %s\n", dep.Repository, dest)
// 	Info("Git: ")
// 	out, err := exec.Command("git", "clone", dep.Repository, dest).CombinedOutput()
// 	fmt.Print(string(out))
// 	return err
// }

// func (g *GitVCS) Update(dep *Dependency) error {
// 	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)

// 	if _, err := os.Stat(dest); err != nil {
// 		Info("Looks like %s is a new package. Cloning.\n", dep.Name)
// 		return g.Get(dep)
// 	}

// 	oldDir, err := os.Getwd()
// 	if err != nil {
// 		return err
// 	}
// 	if err := os.Chdir(dest); err != nil {
// 		return err
// 	}

// 	defer os.Chdir(oldDir)

// 	if oldRepo, err := g.currentRepository(dest); err != nil || oldRepo != dep.Repository {
// 		switch err {
// 		case WrongVCS:
// 			Info("VCS type changed ('%s'). I'm doing a fresh clone.\n", err)
// 		case nil:
// 			Info("Repository changed from %s to %s. I'm doing a clean checkout.\n", oldRepo, dep.Repository)
// 		default:
// 			Info("Unable to determine currently checkout out repository ('%s'). I'm doing a fresh clone.\n", err)
// 		}
// 		os.Chdir(oldDir)
// 		if err := os.RemoveAll(dest); err != nil {
// 			return err
// 		}
// 		return g.Get(dep)
// 	}

// 	if g.currentCheckoutMatchesTagOrRef(dep.Reference) {
// 		Info("%s is up to date.\n", dep.Name)
// 		return nil
// 	}

// 	// Because we can't predict which branch we want to be on, and since
// 	// we want to set checkouts explicitly, we should probably fetch.
// 	//out, err :=  exec.Command("git", "pull", "--ff-only").CombinedOutput()
// 	Info("Git: ")
// 	out, err := exec.Command("git", "fetch", "--all").CombinedOutput()
// 	fmt.Print(string(out))
// 	return err
// }

// func (g *GitVCS) Version(dep *Dependency) error {
// 	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)

// 	oldDir, err := os.Getwd()
// 	if err != nil {
// 		return err
// 	}
// 	os.Chdir(dest)
// 	defer os.Chdir(oldDir)

// 	//Info("Setting %s with 'git checkout'\n", dep.Name)

// 	// Now get to the right reference.
// 	if out, err := exec.Command("git", "checkout", dep.Reference).CombinedOutput(); err != nil {
// 		fmt.Println(string(out))
// 		return err
// 	} else {
// 		updatedTo := "the latest"
// 		if dep.Reference != "" {
// 			updatedTo = dep.Reference
// 		}
// 		Info("Setting version of %s to %s\n", dep.Name, updatedTo)
// 		//fmt.Print(string(out))
// 	}

// 	// EXPERIMENTAL: This will keep the repo up to date according to the
// 	// master branch on Git. Since 'master' is convention only, this isn't
// 	// a 100% reliable way to do things.
// 	if dep.Reference == "" {
// 		Info("No Git reference set. Trying to update 'master'...\n")
// 		dep.Reference = "master"
// 	}

// 	branchref := fmt.Sprintf("origin/%s", dep.Reference)
// 	//err = exec.Command("git", "show-ref", "-q", branchref).Run()
// 	out, err := exec.Command("git", "show-ref", branchref).CombinedOutput()
// 	if err == nil {
// 		//Info("Git: Found branch %s", string(out))
// 		//Debug("Reference %s is to a branch.", dep.Reference)
// 		// git merge --ff-only origin $VERSION
// 		out, err := exec.Command("git", "pull", "--ff-only", "origin", dep.Reference).CombinedOutput()
// 		Info("Git: (merge) %s", string(out))
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	// EXPERIMENTAL: Show how far behind we are.
// 	out, err = exec.Command("git", "rev-list", "--count", "HEAD..origin").CombinedOutput()
// 	if err == nil {
// 		count := strings.TrimSpace(string(out))
// 		if count != "0" {
// 			var c string
// 			switch len(count) {
// 			// 0-9, not that bad
// 			case 1:
// 				c = Green
// 			// 10-99, we're getting behind
// 			case 2:
// 				c = Yellow
// 			// Whoa! We're falling way behind!
// 			default:
// 				c = Red
// 			}
// 			Info(Color(c, fmt.Sprintf("Git: %s is %s behind origin.\n", dep.Name, count)))
// 		}
// 	}

// 	return nil
// }

// func (g *GitVCS) LastCommit(dep *Dependency) (string, error) {
// 	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)

// 	oldDir, err := os.Getwd()
// 	if err != nil {
// 		return "", err
// 	}
// 	os.Chdir(dest)
// 	defer os.Chdir(oldDir)

// 	out, err := exec.Command("git", "log", "-n", "1", "--pretty=format:%h%d").CombinedOutput()
// 	if err != nil {
// 		return "", err
// 	}
// 	parts := strings.SplitN(string(out), " ", 2)

// 	sha := parts[0]

// 	// Send back a tag if a tag matches.
// 	if len(parts) > 1 && strings.Contains(parts[1], "tag: ") {
// 		re := regexp.MustCompile("tag: ([^,)]*)")
// 		subs := re.FindStringSubmatch(parts[1])
// 		if len(subs) > 1 {
// 			return subs[1], nil
// 		}
// 	}

// 	return sha, nil
// }
