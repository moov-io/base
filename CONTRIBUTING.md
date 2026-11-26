# Contributing

Wow, we really appreciate that you even looked at this section! We are trying to make the worlds best atomic building blocks for financial services that accelerate innovation in banking and we need your help!

You only have a fresh set of eyes once! The easiest way to contribute is to give feedback on the documentation that you are reading right now. This can be as simple as sending a message in [Slack](https://slack.moov.io/) with your feedback or updating the markdown in this documentation and issuing a pull request.

Stability is the hallmark of any good software. If you find an edge case that isn't handled please open a GitHub issue with the example data so that we can make our software more robust for everyone. We also welcome pull requests if you want to get your hands dirty.

Have a use case that we don't handle; or handle well! Start the discussion in Slack or open a GitHub Issue. We want to make the project meet the needs of the community and keeps you using our code.

Please review our [Code of Conduct](CODE_OF_CONDUCT.md) to ensure you agree with the values of this project.

We use GitHub to manage reviews of pull requests.

* If you have a trivial fix or improvement, go ahead and create a pull
  request.

* If you plan to do something more involved, first propose your ideas
  in a Github issue. This will avoid unnecessary work and surely give
  you and us a good deal of inspiration.

* Relevant coding style guidelines are the [Go Code Review
  Comments](https://go.dev/wiki/CodeReviewComments)
  and the _Formatting and style_ section of Peter Bourgon's [Go: Best
  Practices for Production
  Environments](https://peter.bourgon.org/go-in-production/#formatting-and-style).

* When in doubt follow the [Go Proverbs](https://go-proverbs.github.io/)

## Getting the code

This project uses [Go Modules](https://go.dev/blog/using-go-modules). You can clone the repository outside of GOPATH.

```
git clone https://github.com/moov-io/base.git
cd base
```

Then add your (or another user's) fork:

```
git remote add $user https://github.com/$user/base.git
git fetch $user
```

Now, feel free to branch and push (`git push $user $branch`) to your remote and send us Pull Requests!

## Tests and checks

- `go test -short ./...` runs unit tests and skips tests that need MySQL, PostgreSQL or Spanner.
- `make setup` starts the test databases. Then `go test ./...` runs the integration tests. `make teardown` stops the containers.
- `make check` runs the project linters (on Windows it runs `go test ./... -short`).

CI runs on GitHub Actions. See the [Go Report Card](https://goreportcard.com/report/github.com/moov-io/base) for lint results on `master`.

## Pull Requests

A good quality PR will have the following characteristics:

* It will be a complete piece of work that adds value in some way.
* It will have a title that reflects the work within, and a summary that helps to understand the context of the change.
* There will be well written commit messages, with well crafted commits that tell the story of the development of this work.
* Ideally it will be small and easy to understand. Single commit PRs are usually easy to submit, review, and merge.
* The code contained within will meet the best practices set by the team wherever possible.
* The code is able to be merged.
* A PR does not end at submission though. A code change is not made until it is merged and used in production.

A good PR should be able to flow through a peer review system easily and quickly.
