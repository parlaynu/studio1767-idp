# Studio 1767 - Prototype OIDC/IdP Server

This project was motivated by a desire to have a common technique to secure internal web APIs for both 
interactive and batch client aplications. 

To do this, it provides a very simple OIDC IdP server that supports authentication using mTLS as well 
as username/password.

Once the client is authenticated, the API servers then work the same for both interactive clients and batch clients.

One of many advantages of using certificates and mTLS over token based authentication is that the client 
certificates can also be imported into the browser (once correctly formatted) and the user can then 
experience transparent authentication in an interactive context.

Note that this is not in a production ready state, and is useful only for experimentation.


## Quick Start

The steps below outline the process to build a simple test environment.

### Build the Test Environment

The below uses terraform to :

* build a local certificate authority
* issue server and user certificates
* create configurations for the Idp and test api server

Run the following commands:

    cd test/build-local
    terraform init
    terraform apply

### Build and Run the IdP Server

    cd cmd/server
    go build
    ./server ../../test/build-local/local/configs/config-idp.yaml

### Build and Run the API Server

    cd test/test-server/cmd/server
    go build
    ./server ./server ../../../build-local/local/configs/config-test-server.yaml

### Run a Test

You can run the test-client application like this:

    cd test/test-client
    python3 -m venv pyenv
    source pyenv/bin/activate
    pip3 install -r requirements.txt
    ./test.py ../build-local/local/configs/certs https://127.0.0.1:8000

You can also start a web browser and browse to `https://127.0.0.1:8000`. This uses a local
certificate authority so you will need to accept that in your browser.

You will be prompted for a username/password to authenticate. If you haven't changed the 
`build-local` project, use 'user2' and 'password2'.


## Testing

The development for this project started with the aws cognito environment and building a test server to work with it.
Then that test server was used to guide the development of this IdP server.

The resulting tools in the `test` directory, and described in the table below, support building test 
environments and running tests.

<table>
  <tr>
    <th>Application</th>
    <th>Description</th>
  </tr>
  <tr>
    <td>test-server</td>
    <td>A simple api server that uses OIDC/IdP for authentication</td>
  </tr>
  <tr>
    <td>test-client</td>
    <td>A command line application that runs tests against the api</td>
  </tr>
  <tr>
    <td>build-cognito</td>
    <td>Builds a test environment with aws cognito as the IdP server. This creates the cognito user pools and adds users and groups to it and generates config files and certificates for the test server to work with it.</td>
  </tr>
  <tr>
    <td>build-local</td>
    <td>Creates certificates and configurations for the test server and this prototype IdP server to work together on the same host.</td>
  </tr>
  <tr>
    <td>build-ldap</td>
    <td>Creates an OpenLDAP server in AWS-EC2 and the configs and certs for the IdP and test servers to use it.</td>
  </tr>
</table>


## To Do

There's quite a bit to do to get this to a production ready status, including the below:

* Clustering
* Token renewal
* Key rotation
* Certificate revocation
* Rigorous testing and validation

