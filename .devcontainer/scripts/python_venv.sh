#! /bin/bash 
set -e
set -x

mkdir /python_venv
python3 -m venv /python_venv/venv 
source /python_venv/venv/bin/activate # You can also tell VSCode to use the interpretter in this location 
pip3 install -r requirements.dev.txt 
pip3 install -r requirements.txt 
echo "source /python_venv/venv/bin/activate" >> "$HOME/.bashrc"
