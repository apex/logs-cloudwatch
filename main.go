//go:generate file2go -in template.yml -pkg main

// Package main provides a command-line tool for setting up the CloudWatch Logs integration for Apex Logs.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"

	flag "github.com/integrii/flaggy"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

// version string.
var version string

// Config struct.
type Config struct {
	StackName string
	ProjectID string
	Region    string
	Endpoint  string
	AuthToken string
	Include   []string
	Exclude   []string
	Template  bool
	Create    bool
}

// TemplateConfig struct.
type TemplateConfig struct {
	Groups []string
}

func main() {
	c := Config{
		StackName: "ApexLogs",
	}

	// flags
	flag.SetVersion(version)
	flag.String(&c.ProjectID, "", "project-id", "Apex Logs destination project ID")
	flag.String(&c.AuthToken, "", "auth-token", "Apex Logs authentication token")
	flag.String(&c.Endpoint, "", "endpoint", "Apex Logs integration endpoint URL")
	flag.String(&c.Region, "", "region", "AWS region id")
	flag.String(&c.StackName, "", "stack-name", "AWS CloudFormation stack name")
	flag.StringSlice(&c.Exclude, "", "exclude", "AWS CloudWatch log group filters")
	flag.StringSlice(&c.Include, "", "include", "AWS CloudWatch log group filters")
	flag.Bool(&c.Create, "", "confirm", "Confirm creation of the stack")
	flag.Bool(&c.Template, "", "template", "Output the template and exit")
	flag.Parse()

	// why you no allow required flags :D
	if c.ProjectID == "" {
		log.Fatalf("error: project-id required")
	}

	if c.Endpoint == "" {
		log.Fatalf("error: endpoint required")
	}

	if c.AuthToken == "" {
		log.Fatalf("error: auth-token required")
	}

	if c.Region == "" {
		log.Fatalf("error: region required")
	}

	// fetch log groups
	fmt.Printf("==> Finding log groups\n")
	groups, err := getLogGroups(c.Include, c.Exclude)
	if err != nil {
		log.Fatalf("error fetching log groups: %s", err)
	}

	fmt.Printf("==> Found %d matching log groups:\n\n", len(groups))
	for _, g := range groups {
		fmt.Printf("    %s\n", g)
	}
	fmt.Printf("\n")

	// render template
	tmpl, err := renderTemplate(TemplateConfig{
		Groups: groups,
	})

	if err != nil {
		log.Fatalf("error rendering template: %s", err)
	}

	// output template
	if c.Template {
		fmt.Printf("%s\n", tmpl)
		return
	}

	f, err := ioutil.TempFile(os.TempDir(), "apex-logs-template-*.yml")
	if err != nil {
		log.Fatalf("error creating tmpfile: %s", err)
	}

	_, err = f.WriteString(tmpl)
	if err != nil {
		log.Fatalf("error writing template file: %s", err)
	}

	err = f.Close()
	if err != nil {
		log.Fatalf("error writing template file: %s", err)
	}

	// create stack
	if c.Create {
		fmt.Printf("==> Creating CloudFormation stack\n")
		err = createStack(f.Name(), c)
		if err != nil {
			log.Fatalf("error creating stack: %s", err)
		}
		return
	}

	fmt.Printf("==> Run command again with --confirm to create the stack\n")
}

// createStack creates a CloudFormation stack.
func createStack(tmplPath string, c Config) error {
	cmd := exec.Command("aws", "cloudformation", "create-stack",
		"--stack-name", c.StackName,
		"--region", c.Region,
		"--template-body", "file://"+tmplPath,
		"--capabilities", "CAPABILITY_IAM",
		"--parameters",
		"ParameterKey=Endpoint,ParameterValue="+c.Endpoint,
		"ParameterKey=AuthToken,ParameterValue="+c.AuthToken,
		"ParameterKey=ProjectID,ParameterValue="+c.ProjectID,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// getLogGroups returns all log groups.
func getLogGroups(include, exclude []string) ([]string, error) {
	logs := cloudwatchlogs.New(session.New(aws.NewConfig()))

	var groups []string
	var cursor *string

	for {
		res, err := logs.DescribeLogGroups(&cloudwatchlogs.DescribeLogGroupsInput{
			NextToken: cursor,
		})

		if err != nil {
			return nil, fmt.Errorf("describing log groups: %w", err)
		}

		cursor = res.NextToken

		for _, group := range res.LogGroups {
			name := *group.LogGroupName
			if filter(name, include, exclude) {
				groups = append(groups, name)
			}
		}

		if cursor == nil {
			break
		}
	}

	return groups, nil
}

// renderTemplate returns a rendered CloudFormation template.
func renderTemplate(c TemplateConfig) (string, error) {
	t := template.Must(template.New("template.yml").Parse(string(TemplateYml)))
	var buf bytes.Buffer
	err := t.Execute(&buf, c)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// filter returns true if the string matches the filters.
func filter(s string, include, exclude []string) bool {
	if include != nil {
		return match(s, include) && !match(s, exclude)
	}

	return !match(s, exclude)
}

// match returns true if s matches any of patterns.
func match(s string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}
