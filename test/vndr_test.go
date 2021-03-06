package vndrtest

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const testRepo = "github.com/docker/swarmkit"

func setGopath(cmd *exec.Cmd, gopath string) {
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "GOPATH=") {
			continue
		}
		cmd.Env = append(cmd.Env, env)
	}
	cmd.Env = append(cmd.Env, "GOPATH="+gopath)
}

func TestVndr(t *testing.T) {
	vndrBin, err := exec.LookPath("vndr")
	if err != nil {
		t.Fatal(err)
	}
	tmp, err := ioutil.TempDir("", "test-vndr-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	repoDir := filepath.Join(tmp, "src", testRepo)
	if err := os.MkdirAll(repoDir, 0700); err != nil {
		t.Fatal(err)
	}

	gitCmd := exec.Command("git", "clone", "https://"+testRepo+".git", repoDir)
	out, err := gitCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to clone %s to %s: %v, out: %s", testRepo, repoDir, err, out)
	}
	if err := os.RemoveAll(filepath.Join(repoDir, "vendor")); err != nil {
		t.Fatal(err)
	}

	vndrCmd := exec.Command(vndrBin)
	vndrCmd.Dir = repoDir
	setGopath(vndrCmd, tmp)

	out, err = vndrCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("vndr failed: %v, out: %s", err, out)
	}
	if !bytes.Contains(out, []byte("Success")) {
		t.Fatalf("Output did not report success: %s", out)
	}

	installCmd := exec.Command("go", "install", testRepo)
	setGopath(installCmd, tmp)
	out, err = installCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install %s failed: %v, out: %s", testRepo, err, out)
	}

	// revendor only etcd
	vndrRevendorCmd := exec.Command(vndrBin, "github.com/coreos/etcd")
	vndrRevendorCmd.Dir = repoDir
	setGopath(vndrRevendorCmd, tmp)

	out, err = vndrRevendorCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("vndr failed: %v, out: %s", err, out)
	}
	if !bytes.Contains(out, []byte("Success")) {
		t.Fatalf("Output did not report success: %s", out)
	}
}

func TestVndrInit(t *testing.T) {
	vndrBin, err := exec.LookPath("vndr")
	if err != nil {
		t.Fatal(err)
	}
	tmp, err := ioutil.TempDir("", "test-vndr-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	repoPath := "github.com/LK4D4"
	repoDir := filepath.Join(tmp, "src", repoPath)
	if err := os.MkdirAll(repoDir, 0700); err != nil {
		t.Fatal(err)
	}

	cpCmd := exec.Command("cp", "-r", "./testdata/dumbproject", repoDir)
	out, err := cpCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("cp failed: %v, out: %s", err, out)
	}
	vndrCmd := exec.Command(vndrBin, "init")
	vndrCmd.Dir = filepath.Join(repoDir, "dumbproject")
	setGopath(vndrCmd, tmp)

	out, err = vndrCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("vndr failed: %v, out: %s", err, out)
	}
	if !bytes.Contains(out, []byte("Success")) {
		t.Fatalf("Output did not report success: %s", out)
	}

	pkgPath := filepath.Join(repoPath, "dumbproject")
	installCmd := exec.Command("go", "install", pkgPath)
	setGopath(installCmd, tmp)
	out, err = installCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install %s failed: %v, out: %s", pkgPath, err, out)
	}
	vndr2Cmd := exec.Command(vndrBin, "init")
	vndr2Cmd.Dir = filepath.Join(repoDir, "dumbproject")
	setGopath(vndr2Cmd, tmp)

	out, err = vndr2Cmd.CombinedOutput()
	if err == nil || !bytes.Contains(out, []byte("There must not be")) {
		t.Fatalf("vndr is expected to fail about existing vendor, got %v: %s", err, out)
	}
}
