# Spresm, at all

## Summary

This RFC proposes the development of Spresm, which is a tool for
dealing with packages of configuration.

You can install and update packages from container images, Helm
charts, or git repos. Installing imports and expands the configuration
into your working directory. Updating merges upstream changes with
those you have made locally.

Spresm uses container images for distributing configuration that comes
as code. You can build such containers with `spresm build`, or
construct them with a Dockerfile or a build pack.

As a special case, `spresm` also lets you treat Helm charts as
packages, in the same way as container images.

## Example

```bash
# install an image as a package
app$ spresm import image app:v1.0.0 ./appconfig
Writing to ./appconfig/ ...
./appconfig/Spresmfile
./appconfig/deployment.yaml
./appconfig/service.yaml
./appconfig/ingress.yaml
app$ git add ./appconfig
app$ git commit -m "Initial import of app:v1.0.0"

# edit the expanded configuration
app$ $EDITOR ./appconfig/deployment.yaml
app$ git add ./appconfig/deployment.yaml
app$ git commit -m "Adapt deployment to local config"

# update from upstream
app$ spresm update ./appconfig --version v1.0.1
Updating version in ./appconfig/Spresmfile
Merging ./appconfig/deployment.yaml
app$ git add -u -- ./appconfig
app$ git commit -m "Update to app:v1.0.1"
```

Installing the package runs the image to generate a base config. From
there you can alter the inputs and generate the configuration again,
edit the YAML files, and update the package from upstream.

Usually, of course, building something and using it would happen in
different places. For example, a continuous integration pipeline would
build the configuration image when you pushed to a pull request, then
it would be tested (e.g., its output linted) and the results reported
back. Once merged, the configuration would be built from the main
branch, and run through another battery of tests, before being pushed
to the image repository. An automated system, or a human, can then
update configurations using the package with the new image.

## Motivation

**Why would you want to do this at all?**

There are a plethora of tools for producing configurations of various
kinds -- all with unique delights, quirks, and drawbacks. For example,
in the Kubernetes world you have (in no particular order):

 - Helm
 - Kustomize
 - jk (my own preference, in a modest third place in no particular
   order)
 - ksonnet (now defunct)
 - [kr8](https://kr8.rocks/) (which I found just now, looking for news
   about ksonnet)
 - Pumuli
 - GOTO 10

What these all have in common is that there's an active component to
them -- you run a program to get a result. Actually there's another,
which is astonishingly popular, that should be added to the list:

 - YAML (and other "just data")

Putting your configuration in a container image, even if it's just
YAML, means you can treat all of it uniformly and not care about _how_
it is expressed.

It also means you get a versioned artifact that you can distribute and
otherwise hang release engineering around. You put programs in
container images and run them through continuous integration and so
on, so why not configurations?

**Why container images and not e.g., tarballs?**

Most tooling has an active component to it -- the tool itself. For
example, you need the `helm` executable to realise a Helm chart.

Using container images means the executable can be shipped with its
input. Won't this be wasteful? No, because container images have
shared structure, so you only need to ship layers that have new
content.

**Why not just use Helm?**

Helm has some quirks and limitations, though certainly plenty of
people think it's good enough. You can keep using Helm charts with
Spresm -- the point is that the rest of the system doesn't depend on
you using Helm charts, or any other specific tooling.

However, Spresm also comes with things you don't get with Helm; in
particular, you can modify the expanded chart, and merge your changes
with changes made to the upstream chart when you upgrade it.

**Shouldn't my configuration be in YAMLs in git?**

Yes, your configuration should be in git. But no, your configuration
doesn't need to be in YAMLs -- that's just a lowest common denominator
for both writing by humans, and reading by tooling and configurable
systems.

All the same techniques for programming, like abstraction and
refactoring, apply for writing configuration as much as for other
programs. Spresm bridges between the environment in which you can use
those techniques, and the environment that needs YAMLs.

## Design

### Plumbing

There are two fundamental operations in Spresm:

 - evaluation (expanding a configuration from its container image or
   chart); and,
 - merging newly expanded configuration with changes made previously.

The rest is dressing around those, to make it usable.

#### Evaluation

The mechanism for evaluation is this:

 - A container using the given image is started;
 - `spresm` assembles the input parameters to the package and writes
   it to the container's stdin;
 - the output is printed to stdout as a list of resources, where it is
   parsed by the runtime and written to files.

#### Merging

The merge operation is so that changes made to the YAML files can be
reconciled with a change in the output of the (newer) container
image. This is a three-way merge between:

 a. the downstream changes, that is anything changed in the local
 configuration;
 b. the upstream changes, that is the change in the output of the
 container image;
 c. the common ancestor, that is the configuration as last evaluated.

### User interface

The porcelain commands in Spresm are largely about importing and
updating packages, but include some conveniences for creating
packages, trouble-shooting, and integration with other systems.

#### Main commands

**Import a package into the local configuration**

    spresm import image org/app:v1.0.0 ./app/
    spresm import chart org/app:v1.0.0 ./app/

This fetches anything it needs (i.e., the image or the chart) and
evaluates it, writing the output files to `app/`.

    spresm import git https://github.com/org/app.git/config app/

This fetches the git repository (everything up to `app.git`) and
copies the files from `./config` in the repository into `app/`.

All of these record the provenance of the config in app/, so that
`spresm update` can fetch another version of the package, and
re-evaluate it.

**Update a package to another version**

    spresm update --version v1.0.1 ./app/

This fetches another version of the configuration package, and merges
it with any changes made to the last version. If the upstream is an
image or chart, it is re-run to get the files to merge.

**Update the expanded package based on input values**

    spresm update --input ./app/

This runs and merges the package again, with (possibly) new input
values but without changing the version.

#### Building packages

For convenience in building packages, Spresm has a set of archetypal
images that can be specialised with program code. These are accessible
with the command `spresm build`:

```bash
$ spresm build --archetype=cdk8s-typescript --image app:v1.0.3 ./config/app
```

#### Evaluating a package

The main mode of use is to import and update packages, which entails
their evaluation; however, it's possible to use the plumbing to
explicitly evaluate packages.

You can run a container locally to generate the configuration, either
in a build script, or just to eyeball it:

```bash
$ spresm evaluate --stdout app:v1.0.3 --values local.yaml
```

---

_Describe here the design of the change._

 - _What is the user interface or API? Give examples if you can._
 - _How will it work, technically?_
 _ _What are the corner cases, and how are they covered?_

## Backward compatibility

_Enumerate any backward-compatibility concerns:_

 - _Will people have to change their existing code?_
 - _If they don't, what might happen?_

_If the change will be backward-compatible, explain why._

## Drawbacks and limitations

_What are the drawbacks of adopting this proposal; e.g.,_

 - _Does it require more engineering work from users (for an
   equivalent result)?_
 - _Are there performance concerns?_
 - _Will it close off other possibilities?_
 - _Does it add significant complexity to the runtime or standard library?_
 - _Does it make understanding `jk` harder?_

## Alternatives

_Explain other designs or formulations that were considered (including
doing nothing, if not already covered above), and why the proposed
design is superior._

## Unresolved questions

 - How is this different to ...
   - Draft -- draft delivers code to a cluster, rather than configuration
   - Skaffold
   - Cloud-native Buildpacks
   - Helm
   - kpt
   - cdk8s

_Keep track here of questions that come up while this is a draft.
Ideally, there will be nothing unresolved by the time the RFC is
accepted. It is OK to resolve a question by explaining why it
does not need to be answered_ yet _._
