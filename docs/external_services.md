# External services

From the Self-hosts' perspective, all services not part of the Self-host itself is an external service. Thus, it doesn't matter if one hosts the service on the same server, on a private but different network in a data centre, or the internet.

The Self-host does not come with any existing external services; we expect third-party developers to create different services as needed or use existing REST services with the help of the [Program Manager and Worker][1].

# External service interaction

An external service may, at its leisure, perform interactions with the Self-host API. For example, it can query objects, add new elements and deleting others. There is nothing in the design that forbids such an interaction, as it shouldn't be.

However, from a deployment viewpoint, one can argue that having multiple separate services that one controls, each using its unique concoctions of configuration and deployment strategy, is a bad design.

Our response to this is to use Self-host `programs` to trigger external services and `Datasets` as a way to manage parametric configuration centrally.

Note that we don't expect one (rightfully so) to store settings such as `listening interface`, `database passwords` and other secrets necessary for the external service to start in a `dataset`. What should go into a `dataset` is `execution` specific arguments, such as Self-host `time-series UUID's, `things UUIDs`, `period ranges` etc. Configurable parameters used as input when triggering an external service.

For example, instead of storing program-specific settings in a program, a dataset can be used, which the program loads on each execution. It then performs its task based on that input. It is possible that another dataset could be referenced by the first as input to an external service, something the program would use as part of the POST data to the external service.

![External service interaction example][fig1]

How interaction with external services takes places depends on your specific needs, the limitation of the platform you have chosen and the rules of the external service you are using.

When developing external services, we recommend that you adhere to the structure outlined in this section. It makes interaction from or to the Self-host API much more effortless.


# References

[1]: <https://github.com/self-host/self-host/blob/main/docs/program_manager_worker.md> "Program Manager and Worker" 
[fig1]: https://raw.githubusercontent.com/self-host/self-host/main/docs/assets/external_service_interaction_ex.svg "External service interaction example"
