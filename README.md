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
  - [7. Versioning](#Versioning)
  - [8. License](#License)
  - [9. Contact Information](#ContactInformation)

## <a name="Description" /> 1. Description

ElasticUtil is a generic library that assists in the use of Elasticsearch, using olivere/elastic library and the offical library. It is possible to create elastic's queries using only one struct. It is also possible to translate errors and responses.

## <a name="TechnologyStack" /> 2. Technology Stack

| **Stack**     | **Version** |
|---------------|-------------|
| Golang        | v1.21.3       |
| golangci-lint | v1.54.2     |

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

- ### <a name="Examples" /> Examples
  
  - [Olivere](https://github.com/arquivei/elasticutil/blob/master/examples/olivere/main.go)

  - [Official library](https://github.com/arquivei/elasticutil/blob/master/examples/official/main.go)

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

## <a name="Versioning" /> 7. Versioning

  We use [Semantic Versioning](http://semver.org/) for versioning. For the versions
  available, see the [tags on this repository](https://github.com/arquivei/gomsgprocessor/tags).

## <a name="License" /> 8. License
  
This project is licensed under the BSD 3-Clause - see the [LICENSE.md](LICENSE.md) file for details.

## <a name="ContactInformation" /> 8. Contact Information

  All contact may be doing by [marcos.filho@arquivei.com.br](mailto:marcos.filho@arquivei.com.br)
