This is a Go application that is primarily deployed through Docker.  The main.go is in the root of the repository and all of the internal implementation details are in the internal/ directory.

## Documentation

The main README.md file provides a minimal overview of the project.

The majority of the documentation is contained within the docs/ directory with separate markdown files for different aspects of the project including one for each command as well as one for the database design, and additional files for each major component. When making changes to the codebase, ensure that the documentation is updated as well.

## Automated Testing

In addition to unit tests that do not interact with external dependencies, integration tests are to be performed by the agent when the agent is testing and debugging functionality.

- GCP Pub/Sub can be emulated using the gcloud with `gcr.io/google.com/cloudsdktool/google-cloud-cli`
- PostgreSQL can be run in a container using the `postgres:18` image from Docker Hub.
