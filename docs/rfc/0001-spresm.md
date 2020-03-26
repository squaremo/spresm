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
back. Once merged, it would be built again and run through another
battery of tests, before being pushed to the image repository and
released into the cluster.

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
think it's good enough. You can still use Helm charts -- the point of
Spresm is so that the rest of the system doesn't depend on you using
Helm charts, or any other specific tooling.

**Shouldn't my configuration be in git?**

Yes, your configuration should be in git. Like all your
code. Packaging it into images is a means of delivery, not a change of
methodology.

## Design

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

_Keep track here of questions that come up while this is a draft.
Ideally, there will be nothing unresolved by the time the RFC is
accepted. It is OK to resolve a question by explaining why it
does not need to be answered_ yet _._

