package drive

import "os/exec"

func ToDefaultTransformation(source string, target string) *exec.Cmd {
	return exec.Command("nice","cpulimit","-l","25","-f","--","ffmpeg","-threads","1", "-i", source, "-y", target)
}
