# ElasticUtil

### A Golang library that assists in the use of Elasticsearch

---------------------

## Table of Contents

  - [1. Description](#Description)
  - [2. Technology Stack](#TechnologyStack)
  - [3. Getting Started](#GettingStarted)
  - [4. Changelog](#Changelog)
  - [5. Collaborators](#Collaborators)
  - [6. Contributing](#Contributing)
  - [7. License](#License)
  - [8. Contact Information](#ContactInformation)

## <a name="Description" /> 1. Description

ElasticUtil is a generic library that assists in the use of Elasticsearch, using olivere/elastic library. It is possible to create olivere/elastic's queries using only one struct. It is also possible to translate errors and responses.

## <a name="TechnologyStack" /> 2. Technology Stack

| **Stack**     | **Version** |
|---------------|-------------|
| Golang        | v1.18       |
| golangci-lint | v1.46.2     |

## <a name="GettingStarted" /> 3. Getting Started

- ### <a name="Prerequisites" /> Prerequisites

  - Any [Golang](https://go.dev/doc/install) programming language version installed, preferred 1.18 or later.

- ### <a name="Install" /> Install
  
  ```
  go get -u github.com/arquivei/elasticutil
  ```

- ### <a name="ConfigurationSetup" /> Configuration Setup

  ```
  go mod vendor
  go mod tidy
  ```

- ### <a name="Usage" /> Usage
  
  - Import the package

    ```go
    import (
        "github.com/arquivei/elasticutil"
    )
    ```
    <!-- TODO -->

- ### <a name="Examples" /> Examples
  
  - [Sample usage](https://github.com/arquivei/elasticutil/blob/master/examples/main.go)

## <a name="Changelog" /> 4. Changelog

  - **ElasticUtil 0.1.0 (May 27, 2022)**
  
    - [New] Decoupling this package from Arquivei's API projects.
    - [New] Setting github's workflow with golangci-lint 
    - [New] Example for usage.
    - [New] Documents: Code of Conduct, Contributing, License and Readme.

## <a name="Collaborators" /> 5. Collaborators

- ### <a name="Authors" /> Authors

  <!-- markdownlint-disable -->
  <!-- prettier-ignore-start -->
	<table>
	<tr>
		<td align="center"><a href="https://github.com/victormn"><img src="https://avatars.githubusercontent.com/u/9757545?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Victor Nunes</b></sub></a></td>
		<td align="center"><a href="https://github.com/rjfonseca"><img src="https://avatars.githubusercontent.com/u/151265?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Rodrigo Fonseca</b></sub></a></td>
	</tr>
	</table>
  <!-- markdownlint-restore -->
  <!-- prettier-ignore-end -->

- ### <a name="Maintainers" /> Maintainers
  
  <!-- markdownlint-disable -->
  <!-- prettier-ignore-start -->
	<table>
	<tr>
		<td align="center"><a href="https://github.com/marcosbmf"><img src="https://avatars.githubusercontent.com/u/34271729?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Marcos Barros</b></sub></a></td>
	</tr>
	</table>
  <!-- markdownlint-restore -->
  <!-- prettier-ignore-end -->

## <a name="Contributing" /> 6. Contributing

  Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## <a name="License" /> 7. License
  
  We use [Semantic Versioning](http://semver.org/) for versioning. For the versions
  available, see the [tags on this repository](https://github.com/arquivei/gomsgprocessor/tags).

## <a name="ContactInformation" /> 8. Contact Information

  All contact may be doing by [marcos.filho@arquivei.com.br](mailto:marcos.filho@arquivei.com.br)
