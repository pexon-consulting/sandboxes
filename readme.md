
# Welcome to your CDK Python project!

This is a blank project for Python development with CDK.

The `cdk.json` file tells the CDK Toolkit how to execute your app.

This project is set up like a standard Python project.  The initialization
process also creates a virtualenv within this project, stored under the `.venv`
directory.  To create the virtualenv it assumes that there is a `python3`
(or `python` for Windows) executable in your path with access to the `venv`
package. If for any reason the automatic creation of the virtualenv fails,
you can create the virtualenv manually.

To manually create a virtualenv on MacOS and Linux:

```
$ python3 -m venv .venv
```

After the init process completes and the virtualenv is created, you can use the following
step to activate your virtualenv.

```
$ source .venv/bin/activate
```

If you are a Windows platform, you would activate the virtualenv like this:

```
% .venv\Scripts\activate.bat
```

Once the virtualenv is activated, you can install the required dependencies.

```
$ pip install -r requirements.txt
```

At this point you can now synthesize the CloudFormation template for this code.

```
$ cdk synth
```

To add additional dependencies, for example other CDK libraries, just add
them to your `setup.py` file and rerun the `pip install -r requirements.txt`
command.

## Useful commands

 * `cdk ls`          list all stacks in the app
 * `cdk synth`       emits the synthesized CloudFormation template
 * `cdk deploy`      deploy this stack to your default AWS account/region
 * `cdk diff`        compare deployed stack with current state
 * `cdk docs`        open CDK documentation

## Local Starten

nur die Test starten oder einen Mock-Server starten

für Test muss du "nur"
`go test ./...` oder wenn du was spezielles Testen willst `go test -v -run TestDeallocateSandbox ./resolver -count=1`
Ich starte die Test/Debug immer über VS-Code

Falls du einen "Server" starten willst musst du im Main.go die funcktion local einkommentieren und in func main =>

```
func main() {
    local()
    // lambda.Start(Handler)
}
```

local aufrufen und das Lambda dings raus nehmen.

`go run main.go -dev`

Grundsätzlich bevor es los geht muss das schema.graphql in go-code kompiliert werden `go generate ./schema` das erzeugt dann ein .go file von wo go dann das schema lesen kann

Enjoy!
