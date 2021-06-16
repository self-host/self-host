# Self-host: Program Manager and Workers

The Program Manager is a required component in a typical deployment. Its purpose is to act as a controller for all Self-host domains and assigning tasks to any number of Program Workers.

If this sounds like any number of already existing software deployment systems, such as Docker och Kubernetes, let's be clear; this solution will NOT replace any of those solutions.

This software is written with the sole intent to provide an easy way to deploy small bits of code without having to do a project out of it and without having to have deep knowledge of either Docker or Kubernetes.


# How it Works

Two components are working together, one Program Manager (juvuln) and one or several instances of the Program Worker (malgomaj).

At set intervals, the Program Manager queries all domains for all active programs. Based on the program's specification, it tracks them and then tells a Program Worker to execute the program. The Program Worker compiles the program (if required), caches the binary and runs the program.

![Interaction Program Manager and Worker][InteractionDiag1]


# Programs and Domains

Programs are declared on a Domain level. This means that there is no way to share programs between domains. It is by design as we wanted to keep hard isolation between different domains.


# Program Languages

There is currently only support for one language; [Tengo](https://github.com/d5/tengo).

We considered support for other languages in the design process, and the API interface between the Program Worker and Program Manager does take program language into account.


# Different types of Programs

There are three different variants of a program.

- module
- routine
- webhook


## Modules
A `module` is a program with the sole purpose of providing support for other programs. In Tengo, to import a module, do;

```golang
mymod := import("mymod")
```

You can use an imported module throughout the program is of it was part of the program.


## Routines
A `routine` is a typical "program". It executes at a given `interval`, e.g. every five minutes. It has a `deadline` that ensures that the Worker will kill the program if it takes too long to execute.

It is an excellent practice to place code that is the same in several different `routines` into a `module`, making the `routine` program smaller.


## Webhooks
A `webhook` is a particular type of program that differs from a `routine` by not executing at a set interval. Instead, it runs upon a call to a webhook endpoint on the Self-host API server.

Internally this is similar (yet not-similar-at-all) to how CGI binaries worker "back in the day".

![Interaction Webhook][InteractionDiag2]


# Allowed Tengo (core) Modules

While Tengo does have several essential core modules, it lacks a module to perform HTTP requests. At compile time of a Tengo program, we make the following modifications to the set of core modules;

1) We add a new "http" module.
2) We add a new "log" module.
3) We remove the "fmt" module and replaces it with our own. We do this to prevent Tengo programs from performing various "print" commands that would show up in stdout.
4) We remove the "os" module as absolute power corrupts absolutely.

Allowed Tengo modules are thus;

- base64 (core)
- enum (core)
- hex (core)
- http (extended)
- json (core)
- log (extended)
- math (core)
- rand (core)
- selfapi (planned for the future)
- text (core)
- times (core)

For more details about Tengo modules, see the official [documentation](https://github.com/d5/tengo/blob/master/docs/stdlib.md).

For details about the extended modules, see our documentation [here](extendedlibs.md).


# Program example

## Simple HTTP request

Let's say you want to deploy a simple program that will execute on the hour of every hour. Its task is to fetch the content from some webpage on the Internet. Let's say the weather forecast for a specific location.

```golang
http := import("http")

MET_NO_URL := "https://api.met.no/weatherapi/locationforecast/2.0/complete.json"
lat := 56.1569464184167
lon := 15.590792749446742

query_args := [
    "lat=" + string(lat),
    "lon=" + string(lon)
]

headers = {
    "User-Agent": "Selfhost Forecast Getter/1.0"
}

response := http.get(
    MET_NO_URL,
    query_args,
    headers
)

// response.Header (is a map[string][]string)

if response.StatusCode == 200 {
    // OK
}

if response.ContentLength > 0 {
    // OK
}

obj := json.decode(response.Body)
// structure of obj depends on the response.
```

The example is nothing more than an example of what an HTTP request could look like. It is up to you to invent something useful with it.


## Named modules from the Self-host API

```golang
mymodule := import("mymodule@latest")
simplemath := import("simplemath@4")

x := simplemath.add(1, 5)
```

Modules declared via the Self-host API is revision based. That means that each new commit of source code represents a new revision (integer). Before anyone can use a code revision, it first has to be signed. Un-signed code will not work as the `Library Manager` (Program Manager) will ignore such revisions.

To always use the latest version of the code for a module, suffix it with `@latest`. To use a specific revision, use `@X` where X is a positive integer.


[InteractionDiag1]: assets/diag_interaction_pmgr_pwrk.svg "Interaction Program Manager and Worker"
[InteractionDiag2]: assets/diag_interaction_webhook.svg "Interaction Webhook"
