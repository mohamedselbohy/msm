#!/bin/bash
cd /tmp
wget -O msm_amd64.deb https://github.com/mohamedselbohy/msm/releases/latest/download/msm_amd64.deb
sudo apt install -y ./msm_amd64.deb
rm msm_amd64.deb
