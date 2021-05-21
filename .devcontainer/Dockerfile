# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.177.0/containers/go/.devcontainer/base.Dockerfile

# [Choice] Go version: 1, 1.16, 1.15
ARG VARIANT="1.15"
FROM mcr.microsoft.com/vscode/devcontainers/go:0-${VARIANT}

# [Option] Install Node.js
ARG INSTALL_NODE="true"
ARG NODE_VERSION="lts/*"
RUN if [ "${INSTALL_NODE}" = "true" ]; then su vscode -c "umask 0002 && . /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION} 2>&1"; fi

# [Optional] Uncomment this section to install additional OS packages.
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    # Install python
    && apt-get -y install --no-install-recommends python python3-dev python3-pip python3-venv python3-setuptools python3-wheel\
    # Install pylint
    && pip3 --disable-pip-version-check --no-cache-dir install pylint

RUN curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - && \ 
    echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | tee -a /etc/apt/sources.list.d/kubernetes.list && \ 
    curl -fsSL https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]')/gpg | (OUT=$(apt-key add - 2>&1) || echo $OUT) && \ 
    echo "deb [arch=amd64] https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]') $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list && \ 
    curl -sL https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor | tee /etc/apt/trusted.gpg.d/microsoft.gpg > /dev/null && \ 
    echo "deb [arch=amd64] https://packages.microsoft.com/repos/azure-cli/ $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/azure-cli.list

RUN apt-get update && apt-get install -y kubectl docker-ce-cli azure-cli

RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 && \
    chmod 700 get_helm.sh && ./get_helm.sh

RUN curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH) && \
    chmod +x kubebuilder && mv kubebuilder /usr/local/bin/

RUN curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.10.0/kind-linux-amd64 && \
    chmod +x ./kind && mv ./kind /bin/kind

RUN curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash

# Install Go tools
RUN \  
    # --> Install gocode
    go get -u github.com/nsf/gocode \ 
    # --> Install go-symbols
    && go get github.com/acroca/go-symbols \ 
    # --> Install impl
    && go get github.com/sasha-s/goimpl/cmd/goimpl \ 
    # --> Install test deps
    && go get github.com/jstemmer/go-junit-report \ 
    && go get github.com/axw/gocov/gocov \ 
	&& go get github.com/AlekSi/gocov-xml \ 
	&& go get github.com/onsi/ginkgo/ginkgo \ 
	&& go get golang.org/x/tools/cmd/cover \ 
	&& go get -u github.com/matm/gocov-html


# tidy up
RUN rm -rf /go/src/ && rm -rf /go/pkg
RUN apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/* 

ENV DATABRICKS_HOST ""
ENV DATABRICKS_TOKEN ""

RUN SNIPPET="export PROMPT_COMMAND='history -a' && export HISTFILE=/commandhistory/.bash_history" && \ 
    echo $SNIPPET >> "/root/.bashrc"
    
# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1