package main

import (
	"fmt"
	"os/exec"

	"github.com/mohamedselbohy/msm/cmd"
)

func main() {
	xhostCmd := exec.Command("xhost", "+local:docker")
	output, err := xhostCmd.CombinedOutput()
	if err != nil {
		fmt.Println("Failed to allow docker to use GUI:", err)
		fmt.Println(string(output))
		return
	}
	cmd.Execute()
}
