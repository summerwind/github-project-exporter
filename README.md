# github-project-exporter

Export GitHub project status to Prometheus.

## Install

### Just want the binary?

Go to the [releases page](https://github.com/summerwind/github-project-exporter/releases), find the version you want, and download the tarball file.

### Run as container?

```
$ docker pull summerwind/github-project-exporter:latest
```

### Building binary yourself

To build the binary you need to install [Go](https://golang.org/), [dep](https://github.com/golang/dep) and [task](https://github.com/go-task/task).

```
$ task vendor
$ task build
```

## Usage

You can specify the repository name with `--github.repository` flag to export the project metrics of the repository. An GitHub access token must be specified in `--github.token`.

```
$ github-project-exporter --github.token ${ACCESS_TOKEN} --github.repository summerwind/github-project-exporter
```

If you want to expose the project metrics of the organization, you can use `--github.organization` instead of `--github.repository`.

```
$ github-project-exporter --github.token ${ACCESS_TOKEN} --github.organization summerwind
```

Multiple flags can also be used.

```
$ github-project-exporter --github.token ${ACCESS_TOKEN} --github.repository summerwind/github-project-exporter --github.repository summerwind/h2spec
```

## Metrics

This exporter will expose these metrics.

| Metric | Meaning | Labels |
| --- | --- | --- |
| github_organization_projects | How many projects are in the organization. | organization |
| github_organization_project_columns | How many columns are in the organization project. | organization, project |
| github_organization_project_cards | How many cards are in the organization project. | organization, project, column |
| github_repository_projects | How many projects are in the repository. | repository |
| github_repository_project_columns | How many columns are in the repository project. | repository, project |
| github_repository_project_cards | How many cards are in the repository project. | repository, project, column |

