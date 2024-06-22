# Github-Actions_Multirun
The Github-Actions_Multirun tool is a Go-based utility designed to streamline the execution of multiple GitHub Actions across various repositories. It allows users to run tests and other workflows for cpecific branch in a microservice architecture setup from a single location, and it presents the results in a clear and concise table.

## Features
- Run Multiple GitHub Actions: Execute GitHub Actions across multiple repositories simultaneously.
- Monitor Workflow Completion: The tool waits for all actions to complete before proceeding.
- Result Aggregation: Collect and display the results of all workflows in a neatly formatted table.
- Configuration Flexibility: Easily configure the actions and other parameters through a YAML configuration file.
- Graceful shutdown: all child workflows will cancelled after configured timeout or after cancellation of main flow.

## Why Use This Tool?
In a microservice architecture, managing multiple repositories can be cumbersome, especially when you need to run tests or other workflows across all of them. GitHub does not natively support running actions across multiple repositories from a single entry point. This tool fills that gap by providing a centralized solution to manage and execute workflows, making it an invaluable asset for developers and DevOps engineers.

## Prerequisites
- Go
- GitHub repositories with workflows configured for repository dispatch events

## Configuration
The workflow_manager_config.yaml file should be created in the project root directory. It includes details about the workflows to be executed and various parameters that control the behavior of the Workflow Manager Tool. Below is an example configuration file structured according to the Go structs.

```
workers_count: 5 # Number of concurrent workers
max_failed_check_attempts: 3 # Maximum number of failed attempts to check workflow status
print_results_interval: 10s # Interval at which to print results
default_initial_check_interval: 30s # Default interval for initial workflow status checks
default_final_check_interval: 1m # Default interval for final workflow status checks
default_attempts_threshold_for_short_check_interval: 5 # Number of attempts before switching to final check interval
time_to_wait_run_ids: 1m # Time to wait for run IDs after dispatching events
multirun_timeout: 15m # Timeout for all workflows to complete
graceful_shutdown_timeout: 5m # Timeout for graceful shutdown
cancel_interval: 1m # Interval to check for cancellation requests
dispatch_event_type: "repository_dispatch" # Event type for dispatching workflows

workflows:
  - repo_name: "repo1"
    owner_name: "owner1"
    workflow_name: "test-workflow.yml"
    alias: "Test Repo 1"
    initial_check_interval: 20s
    final_check_interval: 1m
    attempts_threshold_for_short_check_interval: 3
    params:
      param1: value1
      param2: value2

  - repo_name: "repo2"
    owner_name: "owner2"
    workflow_name: "build-workflow.yml"
    alias: "Build Repo 2"
    initial_check_interval: 30s
    final_check_interval: 2m
    attempts_threshold_for_short_check_interval: 4
    params:
      param1: value1
      param2: value2
```

### Configuration Parameters
- workers_count: The number of concurrent workers to run workflows.
max_failed_check_attempts: Maximum number of attempts to check the status of a workflow before considering it failed.
- print_results_interval: Time interval for printing the workflow results.
default_initial_check_interval: Default interval for the initial checks of workflow status.
- default_final_check_interval: Default interval for the final checks of workflow status.
- default_attempts_threshold_for_short_check_interval: Number of attempts before switching from initial to final check interval.
- time_to_wait_run_ids: Duration to wait for run IDs after dispatching events.
- multirun_timeout: Overall timeout for the execution of all workflows.
graceful_shutdown_timeout: Timeout for graceful shutdown of the tool.
- cancel_interval: Interval to check for any cancellation requests.
dispatch_event_type: Event type used to dispatch workflows in the repositories.

### Workflow Configuration
Each workflow configuration includes the following parameters:

- repo_name: Name of the repository.
owner_name: Owner of the repository.
workflow_name: Name of the workflow file in the repository.
- alias: An alias for easy identification of the workflow.
initial_check_interval: Interval for the initial checks of the workflow status.
- final_check_interval: Interval for the final checks of the workflow status.
attempts_threshold_for_short_check_interval: Number of attempts before - switching from initial to final check interval.
- params: Additional parameters to be passed to the workflow.

Make sure to configure the workflow_manager_config.yaml file according to your specific needs and repository settings.

## Usage
Build the tool with:

```make ./bin/multirun```

Run the tool with the following command:

```./workflow-manager run --params "branch=<your_branch>,some_otherparam=some_value"```

The tool will start executing the workflows as per the configuration and will wait for their completion. Once done, it will print the results in a tabular format.