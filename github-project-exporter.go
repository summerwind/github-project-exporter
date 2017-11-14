package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	VERSION string = "latest"
	COMMIT  string = "HEAD"
)

const (
	NAMESPACE = "github"
)

var (
	descOrgProjects = prometheus.NewDesc(
		prometheus.BuildFQName(NAMESPACE, "", "organization_projects"),
		"How many projects are in the organization.",
		[]string{"organization"},
		nil,
	)

	descOrgProjectColumns = prometheus.NewDesc(
		prometheus.BuildFQName(NAMESPACE, "", "organization_project_columns"),
		"How many columns are in the organization project.",
		[]string{"organization", "project"},
		nil,
	)

	descOrgProjectCards = prometheus.NewDesc(
		prometheus.BuildFQName(NAMESPACE, "", "organization_project_cards"),
		"How many cards are in the organization project.",
		[]string{"organization", "project", "column"},
		nil,
	)

	descRepoProjects = prometheus.NewDesc(
		prometheus.BuildFQName(NAMESPACE, "", "repository_projects"),
		"How many projects are in the repository.",
		[]string{"repository"},
		nil,
	)

	descRepoProjectColumns = prometheus.NewDesc(
		prometheus.BuildFQName(NAMESPACE, "", "repository_project_columns"),
		"How many columns are in the repository project.",
		[]string{"repository", "project"},
		nil,
	)

	descRepoProjectCards = prometheus.NewDesc(
		prometheus.BuildFQName(NAMESPACE, "", "repository_project_cards"),
		"How many cards are in the repository project.",
		[]string{"repository", "project", "column"},
		nil,
	)
)

type Exporter struct {
	client *github.Client
	orgs   []string
	repos  []string
}

func NewExporter(token string, orgs []string, repos []string) (*Exporter, error) {
	if token == "" {
		return nil, errors.New("invalid token")
	}

	if len(orgs) == 0 && len(repos) == 0 {
		return nil, errors.New("at least one organization name or repository name is required")
	}

	if len(repos) > 0 {
		for _, repo := range repos {
			r := strings.Split(repo, "/")
			if len(r) != 2 || (r[0] == "" || r[1] == "") {
				return nil, fmt.Errorf("invalid repository name: %s", repo)
			}
		}
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	hc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(hc)

	return &Exporter{
		client: client,
		orgs:   orgs,
		repos:  repos,
	}, nil
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- descOrgProjects
	ch <- descOrgProjectColumns
	ch <- descOrgProjectCards
	ch <- descRepoProjects
	ch <- descRepoProjectColumns
	ch <- descRepoProjectCards
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	for _, org := range e.orgs {
		projects, err := e.getOrganizationProjects(org)
		if err != nil {
			log.Errorln(err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			descOrgProjects, prometheus.GaugeValue, float64(len(projects)), org,
		)

		for _, project := range projects {
			columns, err := e.getProjectColumns(*project.ID)
			if err != nil {
				log.Errorln(err)
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				descOrgProjectColumns, prometheus.GaugeValue, float64(len(columns)), org, strconv.Itoa(*project.Number),
			)

			for _, column := range columns {
				cards, err := e.getProjectCards(*column.ID)
				if err != nil {
					log.Errorln(err)
					continue
				}

				ch <- prometheus.MustNewConstMetric(
					descOrgProjectCards, prometheus.GaugeValue, float64(len(cards)), org, strconv.Itoa(*project.Number), *column.Name,
				)
			}
		}
	}

	for _, repo := range e.repos {
		projects, err := e.getRepositoryProjects(repo)
		if err != nil {
			log.Errorln(err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			descRepoProjects, prometheus.GaugeValue, float64(len(projects)), repo,
		)

		for _, project := range projects {
			columns, err := e.getProjectColumns(*project.ID)
			if err != nil {
				log.Errorln(err)
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				descRepoProjectColumns, prometheus.GaugeValue, float64(len(columns)), repo, strconv.Itoa(*project.Number),
			)

			for _, column := range columns {
				cards, err := e.getProjectCards(*column.ID)
				if err != nil {
					log.Errorln(err)
					continue
				}

				ch <- prometheus.MustNewConstMetric(
					descRepoProjectCards, prometheus.GaugeValue, float64(len(cards)), repo, strconv.Itoa(*project.Number), *column.Name,
				)
			}
		}
	}
}

func (e *Exporter) getOrganizationProjects(org string) ([]*github.Project, error) {
	var projects []*github.Project

	if org == "" {
		return nil, fmt.Errorf("invalid organization name: %s", org)
	}

	opts := &github.ProjectListOptions{
		State: "open",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	for {
		pjs, res, err := e.client.Organizations.ListProjects(context.Background(), org, opts)
		if err != nil {
			return nil, fmt.Errorf("unable to get organization projects: %s", org)
		}

		projects = append(projects, pjs...)

		if res.NextPage == 0 {
			break
		}
		opts.Page = res.NextPage
	}

	return projects, nil
}

func (e *Exporter) getRepositoryProjects(repo string) ([]*github.Project, error) {
	var projects []*github.Project

	r := strings.Split(repo, "/")
	if len(r) != 2 {
		return nil, fmt.Errorf("invalid repository name: %s", repo)
	}

	opts := &github.ProjectListOptions{
		State: "open",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	for {
		pjs, res, err := e.client.Repositories.ListProjects(context.Background(), r[0], r[1], opts)
		if err != nil {
			return nil, fmt.Errorf("unable to get repository projects: %s/%s", r[0], r[1])
		}

		projects = append(projects, pjs...)

		if res.NextPage == 0 {
			break
		}
		opts.Page = res.NextPage
	}

	return projects, nil
}

func (e *Exporter) getProjectColumns(projectID int) ([]*github.ProjectColumn, error) {
	var columns []*github.ProjectColumn

	opts := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}

	for {
		cols, res, err := e.client.Projects.ListProjectColumns(context.Background(), projectID, opts)
		if err != nil {
			return nil, fmt.Errorf("unable to get project columns: %d", projectID)
		}

		columns = append(columns, cols...)

		if res.NextPage == 0 {
			break
		}
		opts.Page = res.NextPage
	}

	return columns, nil
}

func (e *Exporter) getProjectCards(columnID int) ([]*github.ProjectCard, error) {
	var cards []*github.ProjectCard

	opts := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}

	for {
		cas, res, err := e.client.Projects.ListProjectCards(context.Background(), columnID, opts)
		if err != nil {
			return nil, fmt.Errorf("unable to get project cards: %d", columnID)
		}

		cards = append(cards, cas...)

		if res.NextPage == 0 {
			break
		}
		opts.Page = res.NextPage
	}

	return cards, nil
}

func main() {
	var cmd = &cobra.Command{
		Use:   "github-project-exporter",
		Short: "Exporter for GitHub Project",
		RunE:  run,
	}

	cmd.Flags().String("github.token", "", "GitHub access token")
	cmd.Flags().StringSlice("github.organization", []string{}, "Organization name")
	cmd.Flags().StringSlice("github.repository", []string{}, "Repository name")
	cmd.Flags().String("web.listen-address", "0.0.0.0:9410", "Address to listen on for web interface and telemetry")
	cmd.Flags().String("web.telemetry-path", "/metrics", "Path under which to expose metrics")
	cmd.Flags().Bool("version", false, "Display version information and exit")

	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	v, err := cmd.Flags().GetBool("version")
	if err != nil {
		return err
	}

	if v {
		version()
		os.Exit(0)
	}

	listenAddress, err := cmd.Flags().GetString("web.listen-address")
	if err != nil {
		return err
	}

	telemetryPath, err := cmd.Flags().GetString("web.telemetry-path")
	if err != nil {
		return err
	}

	token, err := cmd.Flags().GetString("github.token")
	if err != nil {
		return err
	}

	orgs, err := cmd.Flags().GetStringSlice("github.organization")
	if err != nil {
		return err
	}

	repos, err := cmd.Flags().GetStringSlice("github.repository")
	if err != nil {
		return err
	}

	exporter, err := NewExporter(token, orgs, repos)
	if err != nil {
		return err
	}
	prometheus.MustRegister(exporter)

	http.Handle(telemetryPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><head>
		<title>GitHub Project Exporter</title></head><body>
		<h1>GitHub Project Exporter</h1>
    <p><a href="` + telemetryPath + `">Metrics</a></p>
    </body></html>`))
	})

	log.Infoln("Listening on", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))

	return nil
}

func version() {
	fmt.Printf("Version: %s (%s)\n", VERSION, COMMIT)
}