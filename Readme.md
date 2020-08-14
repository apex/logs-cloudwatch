
# CloudWatch → Apex Logs 

The `logs-cloudwatch` command-line tool provides an easy way to send AWS CloudWatch Logs to [Apex Logs](https://apex.sh/logs/).

## Installation

```
```

## Usage

Command usage:

```
logs-cloudwatch

  Flags: 
       --version      Displays the program version string.
    -h --help         Displays help with available flag, subcommand, and positional value parameters.
       --project-id   Apex Logs destination project ID
       --auth-token   Apex Logs authentication token
       --endpoint     Apex Logs integration endpoint URL
       --region       AWS region id
       --stack-name   AWS CloudFormation stack name (default: ApexLogs)
       --exclude      AWS CloudWatch log group filters
       --include      AWS CloudWatch log group filters
       --confirm      Confirm creation of the stack (default: false)
       --template     Output the template and exit

```

## Example

Provide your Apex Logs destination project ID, API endpoint (with `/integrations/cloudwatch` as the path), and API token to provide AWS write-access to your logs. For help creating an API token visit the [Apex Logs documentation](https://apex.sh/docs/logs/api/#authentication).

```
logs-cloudwatch \
  --project-id <project-id> \
  --endpoint https://<endpoint>/integrations/cloudwatch \
  --auth-token <api-token> \
  --include /aws/lambda/ \
  --region us-west-2
```

Omitting the `--confirm` flag will output a preview of the matching log groups:

```
==> Finding log groups
==> Found 4 matching log groups:

    /aws/lambda/logs-api
    /aws/lambda/news-api
    /aws/lambda/primary-api
    /aws/lambda/up-api

==> Run command again with --confirm to create the stack
```

Tweak your `--include` and `--exclude` filters as necessary to get the log groups you want, then add the `--confirm` flag to generate the stack.

## Filtering

Running `logs-cloudwatch` without `--include` or `--exclude` flags will subscribe to all log groups. The filter patterns are simply sub-string matches, for example `get_` would match `/aws/lambda/get_team_members`, and `API` would match `API-Gateway-Execution-Logs_g2sdfdwn5rkc6/production`.

### Examples

Send all logs:

```sh
$ logs-cloudwatch
```

Include only AWS Lambda logs, matching groups such as `/aws/lambda/get_team_members`:

```sh
$ logs-cloudwatch --include /aws/lambda/
```

Include only AWS Lambda logs, matching groups such as `/aws/lambda/get_team_members`, but excluding a few.

```sh
$ logs-cloudwatch --include /aws/lambda/ --exclude /aws/lambda/api,/aws/lambda/app
```

Exclude RDS and API Gateway logs:

```sh
$ logs-cloudwatch --exclude API-Gateway,RDS
```

