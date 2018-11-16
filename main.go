package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

// go run main.go run <cmd> <args>
func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("help")
	}
}

func run() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	must(cmd.Run())
}

const root = "/home/centos/mroot"

var toMount = []string{
	"/proc",
	"/dev",
	"/etc",
	"/lib64",
	"/bin",
}

func child() {
	cg()

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//	must(syscall.Sethostname([]byte("container")))

	// make dirs
	os.Mkdir(root, os.ModeDir)
	for _, dir := range toMount {
		os.Mkdir(root+dir, os.ModeDir)
	}

	// mount dirs
	for _, dir := range toMount {
		must(syscall.Mount(dir, root+dir, "", syscall.MS_BIND, ""))
	}

	must(syscall.Chroot(root))
	must(os.Chdir("/"))
	//	must(syscall.Mount("thing", "mytemp", "tmpfs", 0, ""))s

	defer func() {
		for _, dir := range toMount {
			must(syscall.Unmount(dir, 0))
		}
		//		must(syscall.Unmount("mytemp", 0))
	}()

	//	fmt.Println(os.Getenv("PATH"))

	//	files, err := ioutil.ReadDir("/bin")
	//	if err != nil {
	//		panic(err)
	//	}

	//	for _, f := range files {
	//		fmt.Println(f.Name())
	//	}

	must(cmd.Run())
}

func cg() {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	os.Mkdir(filepath.Join(pids, "basher"), 0755)
	must(ioutil.WriteFile(filepath.Join(pids, "basher/pids.max"), []byte("20"), 0700))
	// Removes the new cgroup in place after the container exits
	must(ioutil.WriteFile(filepath.Join(pids, "basher/notify_on_release"), []byte("1"), 0700))
	must(ioutil.WriteFile(filepath.Join(pids, "basher/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
