# MSM - CLI
This CLI tool utilizes the ROS docker image to create and handle ROS workspaces and packages. It can also be used to exec into workspaces, launch packages with dependencies.

> [!NOTE]
> Completions are available and the tool is well documented under `msm -h` or `msm help` or `msm --help`
## Installation
Install the latest *.deb package from the releases by the following commands.
```bash
cd /tmp
wget -O msm_amd64.deb https://github.com/mohamedselbohy/msm/releases/latest/download/msm_amd64.deb
sudo apt install -y ./msm_amd64.deb
rm msm_amd64.deb
```
or run installation script `./install.sh`
## Build Instructions
At the project's root directory just simply execute these commands.
```bash
go build -o msm
sudo ln -s "$(pwd)/msm" /usr/local/bin/msm
```
Then ensure installation with
```bash
msm --help
```


## Autocompletions
Fish: `msm completion fish > ~/.config/fish/completions/msm.fish`

Bash: `sudo bash -c "msm completion bash > /etc/bash_completion.d/msm"`
