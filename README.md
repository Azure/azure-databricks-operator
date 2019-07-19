
[![Build Status](https://img.shields.io/azure-devops/build/ms/c7bb5181-d75b-4ef1-8145-a2c051657858/152.svg?label=build-azure-databricks-operator&style=flat-square&logo=data%3Aimage%2Fpng%3Bbase64%2CiVBORw0KGgoAAAANSUhEUgAAADQAAAAyCAMAAAAk%2FwjEAAAAXVBMVEX%2F%2F%2F%2BTk5Obm5udnZ2kpKSlpaWnp6eoqKiwsLCxsbGysrKzs7O6urq7u7u9vb3GxsbHx8fIyMjJycnR0dHS0tLT09PU1NTd3d3e3t7f39%2Fo6Ojp6enq6ur09PT%2F%2F%2F%2Bel%2BNbAAAAAXRSTlMAQObYZgAAAM5JREFUeNrt090OgjAMhuEqCoqKv8Mprvd%2FmUY6U9SFfSwemMh7toMno82gsd%2BpcNxtgpiMX6sR5BKQ5eGoYOlARDWKvMnawwpDVsyFpBuCSpZyPccRv6EcQA37KpKqONrxM9e52VCgJQdq9OYg4nDWWGYcaemoTECWhqM94UhX%2F21kHbI4zcm0n8pNe%2F9v63F8bZrRl6fNCWgT2UDcsEHMop1C5zpRf%2FL5TiZ3MJq1yzvLCUc6WgJaM47U4Oh6fKRrQJCPumoLI1UNjf1ddw%2FHSv3TGNoxAAAAAElFTkSuQmCC)](https://dev.azure.com/ms/azure-databricks-operator/_build/latest?definitionId=152&branchName=master)


# Azure Databricks operator (for Kubernetes)

> This project is experimental. Expect the API to change. It is not recommended for production environments.

## Introduction

Kubernetes offers the facility of extending it's API through the concept of 'Operators' ([Introducing Operators: Putting Operational Knowledge into Software](https://coreos.com/blog/introducing-operators.html)). This repository contains the resources and code to deploy an Azure Databricks Operator for Kubernetes.

It is a Kubernetes controller that watches Customer Resource Definitions (CRDs) that define a Databricks job.

![alt text](docs/images/azure-databricks-operator.jpg "high level architecture")

The Databricks operator is useful in situations where Kubernetes hosted applications wish to launch and use Databricks data engineering and machine learning tasks.

The project was built using

1. [Kubebuilder](https://book.kubebuilder.io/)
2. [Golang SDK for DataBricks](https://github.com/xinsnake/databricks-sdk-golang)

## Quick start

For deployment gudes please see [deploy.md](https://github.com/microsoft/azure-databricks-operator/blob/master/docs/deploy.md)

## Roadmap

Check [roadmap.md](https://github.com/microsoft/azure-databricks-operator/blob/master/docs/roadmap.md) for what has been supported and what's coming.

## Resources

Few topics are discussed in the [resources.md](https://github.com/microsoft/azure-databricks-operator/blob/master/docs/resources.md)

- Kubernetes on WSL
- Build pipelines

## Contributing

For instructions about setting up your environment to develop and extend the operator, please see
[contributing.md](https://github.com/microsoft/azure-databricks-operator/blob/master/docs/contributing.md)

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.microsoft.com.

When you submit a pull request, a CLA-bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., label, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
