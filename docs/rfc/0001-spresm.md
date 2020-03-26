# Spresm, at all

## Summary

This RFC proposes the development of Spresm, which is a tool for
packaging configuration as container images.

## Example

```bash
app$ # builds an image from the jk code in the current directory, using a buildpack
app$ spresm build --buildpack jk --image app:$(git revparse HEAD) .
Created image app:a79a826287f8c4ea63efbb3a356423e0d6ec6d52
app$ # uses the image to generate the configuration into /tmp/appconfig,
app$ # supplying some configuration
app$ spresm generate app:a79a826287f8c4ea63efbb3a356423e0d6ec6d52 --values=local.yaml -o /tmp/appconfig
Writing to /tmp/appconfig/ ...
/tmp/appconfig/deployment.yaml
/tmp/appconfig/service.yaml
/tmp/appconfig/ingress.yaml
```

Usually, of course, these two things would happen in different places
-- for example, a continuous integration pipeline would build the
configuration image when you pushed to a pull request, then it would
be tested (e.g., its output linted) and the results reported
back. Once merged, the configuration would be built from master, and
run through another battery of tests, before being pushed to the image
repository and released to the running system.

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

Using container images means the executable can come with its
input. But won't this be wasteful? No, because container images have
shared structure, so you only need to ship layers that have new
content.

**Why not just use Helm?**

Helm has some quirks and limitations, but certainly plenty of people
think it's good enough. You can keep using Helm charts with Spresm --
the point is that the rest of the system doesn't depend on you using
Helm charts, or any other specific tooling.

**Shouldn't my configuration be in git?**

Yes, your configuration should be in git. Like all your
code. Packaging it into images is a means of delivery, not a change of
methodology.

## Design

### Mechanism

The basic mechanism for evaluation is this:

 - `spresm` assembles the input parameters to the configuration, from
   the files or individual paramaters given to it;

 - the entrypoint of the container in question is run, mounting an
   output directory, as well as the parameters at a conventional
   location and supplying the output directory as an environment
   variable;

 - when complete, the files in the output directory are used as the
   result.

### Using Spresm locally

You can run a container locally to generate the configuration, either
in a build script, or just to eyeball it:

```bash
$ spresm generate --stdout app:v1.0.3 --values local.yaml
```

### Using Spresm in continuous integration

Spresm might feature in a few places in a continuous integration
pipeline:

 - it could be invoked on a set of parameters particular to the
   environment, then checked with linting and so on, to verify that
   the configuration will work in that environment (e.g., a cluster)

 - it could be used as the last step of a delivery pipeline, to
   generate the configuration that will be applied to a running system

 - the definition of the pipeline might itself be packaged in an
   image, and applied with `spresm`.

### Using Spresm to drive a GitOps pipeline

You often want to see the effect that a configuration change will make
to the running system, and this is usually easier to evaluate as data;
e.g., as YAMLs.

For this purpose, you can keep a git repository of YAML files, and use
`spresm` to splat new configuration into it. That not only gives you
diffs of the data, but means you can use that git repository as the
system of record for GitOps.

### Using Spresm as an operator

Spresm could be tooled as an operator to run in Kubernetes. This would
watch a particular kind of custom resource, which specified the image
version as well as values from various sources, and apply the
configuration generated whenever anything changed.

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

_Keep track here of questions that come up while this is a draft.
Ideally, there will be nothing unresolved by the time the RFC is
accepted. It is OK to resolve a question by explaining why it
does not need to be answered_ yet _._
