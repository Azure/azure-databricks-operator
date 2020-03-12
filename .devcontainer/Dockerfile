#-------------------------------------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation. All rights reserved.
# Licensed under the MIT License. See https://go.microsoft.com/fwlink/?linkid=2090316 for license information.
#-------------------------------------------------------------------------------------------------------------

FROM golang:1.12-buster

ARG KubectlVersion=v1.16.2
ARG HelmVersion=v3.0.3

# Avoid warnings by switching to noninteractive
ENV DEBIAN_FRONTEND=noninteractive

# Configure apt, install packages and tools
RUN apt-get update \
    && apt-get -y install --no-install-recommends apt-utils 2>&1 \
    # Verify git, process tools, lsb-release (common in install instructions for CLIs) installed
    && apt-get -y install git procps lsb-release \
    # Install python
    && apt-get -y install --no-install-recommends git openssl build-essential ca-certificates nano curl python python3-dev python3-pip python3-venv python3-setuptools python3-wheel\
    # Install pylint
    && pip3 --disable-pip-version-check --no-cache-dir install pylint \
    # Install Editor
    && apt-get install vim -y \
    # Clean up
    && apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/* 

RUN apt-get update \
    # Install Docker CE CLI
    && apt-get install -y apt-transport-https ca-certificates curl gnupg-agent software-properties-common lsb-release \
    && curl -fsSL https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]')/gpg | apt-key add - 2>/dev/null \
    && add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]') $(lsb_release -cs) stable" \
    && apt-get update \
    && apt-get install -y docker-ce-cli \
    # Install kubectl
    && curl -sSL -o /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/$KubectlVersion/bin/linux/amd64/kubectl \
    && chmod +x /usr/local/bin/kubectl \
    # Install Helm
    && curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | DESIRED_VERSION=$HelmVersion bash \
    && helm repo add stable https://kubernetes-charts.storage.googleapis.com/ 


# Enable go modules
ENV GO111MODULE=on

# Install Go tools
RUN \
    # --> Delve for debugging
    go get github.com/go-delve/delve/cmd/dlv@v1.3.2 \
    # --> Go language server
    && go get golang.org/x/tools/gopls@v0.2.1 \
    # --> Go symbols and outline for go to symbol support and test support 
    && go get github.com/acroca/go-symbols@v0.1.1 && go get github.com/ramya-rao-a/go-outline@7182a932836a71948db4a81991a494751eccfe77 \
    # --> GolangCI-lint
    && curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sed 's/tar -/tar --no-same-owner -/g' | sh -s -- -b $(go env GOPATH)/bin \
    # --> Install Ginkgo
    && go get github.com/onsi/ginkgo/ginkgo@v1.12.0 \
    # --> Install junit converter
    && go get github.com/jstemmer/go-junit-report@v0.9.1 \
    && rm -rf /go/src/ && rm -rf /go/pkg

# Enable bash completion
RUN apt-get update && apt install -y bash-completion && echo "source /etc/bash_completion" >> "/root/.bashrc"

# Verify git, process tools installed
RUN apt-get -y install git procps wget nano zsh inotify-tools jq
RUN wget https://github.com/robbyrussell/oh-my-zsh/raw/master/tools/install.sh -O - | zsh || true

# Install golangci-linter
RUN curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(go env GOPATH)/bin v1.21.0

ENV PATH="/usr/local/kubebuilder/bin:${PATH}"

# Set the default shell to bash instead of sh
ENV DATABRICKS_HOST ""
ENV DATABRICKS_TOKEN ""

COPY ./Makefile ./
RUN make install-kind
RUN make install-kubebuilder
RUN make install-kustomize
RUN make install-test-dependency
# tidy up go packages
RUN rm -rf /go/src/ && rm -rf /go/pkg

ENV SHELL /bin/bash

# Save command line history 
RUN echo "export HISTFILE=/root/commandhistory/.bash_history" >> "/root/.bashrc" \
    && echo "export PROMPT_COMMAND='history -a'" >> "/root/.bashrc" \
    && mkdir -p /root/commandhistory \
    && touch /root/commandhistory/.bash_history

# Add useful aliases
RUN echo "alias k=kubectl" >> "/root/.bashrc"
# Add autocomplete to kubectl
RUN echo "source <(kubectl completion bash)" >> "/root/.bashrc"
RUN echo "source <(kubectl completion bash | sed 's/kubectl/k/g')" >> "/root/.bashrc"
# Add kubectx 
RUN git clone https://github.com/ahmetb/kubectx.git /root/.kubectx \
    && COMPDIR=$(pkg-config --variable=completionsdir bash-completion) \
    && ln -sf /root/.kubectx/completion/kubens.bash $COMPDIR/kubens \
    && ln -sf /root/.kubectx/completion/kubectx.bash $COMPDIR/kubectx 

# Git command prompt
RUN git clone https://github.com/magicmonty/bash-git-prompt.git ~/.bash-git-prompt --depth=1 \
    && echo "if [ -f \"$HOME/.bash-git-prompt/gitprompt.sh\" ]; then GIT_PROMPT_ONLY_IN_REPO=1 && source $HOME/.bash-git-prompt/gitprompt.sh; fi" >> "/root/.bashrc"

ENV PATH="/root/.kubectx:${PATH}"

COPY ./locust/requirements.* ./
COPY ./.devcontainer/scripts/python_venv.sh ./

RUN   bash -f ./python_venv.sh
