# Authentication (Users)

To handle authentication, the API used both [RFC-7617](https://tools.ietf.org/html/rfc7617) (The 'Basic' HTTP Authentication Scheme) and [RFC-8959](https://tools.ietf.org/html/rfc8959) (The 'secret-token' URI Scheme).

Example;  
`Authorization: Basic c2VjcmV0LXRva2VuOm15ZG9tYWluLllhNGJkNHph`.

The Base64 encoded part in the example above decodes to:  
`mydomain:secret-token.Ya4bd4za`.

The access token is composed of three parts;

* domain (*mydomain*)
* secret-token-scheme (*secret-token*)
* secret (*Ya4bd4za*)

A single colon (":") separates the domain and token parts, and a single dot (".") separates the secret-token-scheme and secret parts. This format effectively makes the domain the username and the "secret-token-scheme.secret" the password from RFC-7617.

The domain part is used to check the secret against the correct domain as each access token is unique to a specific domain. We chose this structure as it is a straightforward extension of a well-established authentication process.

**NOTE**: The value of secret-token-scheme must always be *secret-token*. This value aims to ease automatic identification and prevention of keys in source code.


## Questions and Answers

### Why not use JWT (JSON Web Token)? Is it an IETF standard? Also, it is an excellent fit for REST API:s

While it is true that JWT is a proposed standard by the IETF, and it is also true that it removes the requirement to authenticate each request. JWT carries a significant transport overhead as each request holds a simple token and an entire encrypted blob of meta-data. We think that this overhead outweighs the benefits of having to authenticate each request. Even more so, since the only way to disown a client is to use a global blacklist. As each request has to be validated against this list, we re-introduce the same overhead we have when authenticating each request. This overhead renders JWT pointless from a performance perspective. 