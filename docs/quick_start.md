This chapter explains the complete process of uploading a document.
You need to have completed signup and project creation beforehand.
If you haven't completed these steps yet, please follow the document below to first complete signup and project creation.

https://document.do.dodo-doc.com/install

## Steps to Upload a Document
The general process for uploading a document is as follows:

* Generate a new API Key from the project screen and set it as an environment variable
* Use the `dodo-cli init` command to generate a configuration file
* Use the `dodo-cli upload` command to upload the document

Now, let's delve into the specific steps in more detail.

# Creating an API Key
An API Key is necessary to verify that the user running dodo-cli has the appropriate permissions.
First, let's log into Dodo in your browser and generate a new API Key.

You can generate API Keys from each project's screen.
From the dashboard, click on the project you want to upload to in order to open the project screen.

https://www.dodo-doc.com/dashboard

Clicking the `New API Key` button in the upper right corner of the project screen will generate a new API Key.
The newly issued API Key will be displayed at the top of the screen. Please copy and store it securely.

:::message warning
The API Key is only displayed once and cannot be viewed again after you close the screen.
:::

:::message alert
Do not share your API Key on the internet or in any public space.
If your API Key is leaked, there is a risk that the content of your documents could be tampered with.
:::

# Creating a Configuration File
Next, we'll create a configuration file for dodo.
You can easily create a template for the configuration file using the `dodo-cli init` command.

First, navigate to the root of your git repository and run the dodo-cli command.
You'll be asked a few questions in an interactive format. Please answer them.
Once you've finished answering the questions, a new configuration file named `.dodo.yaml` will be generated.

```yaml
version: 1
project:
  name: testdoc
  version: 1
  description: test description
pages:
  - markdown: README.md
    path: "/README"
    title: "README"
  ## Create the directory and place all markdown files in the docs
  #- directory: "Directory"
  #  children:
  #    - match: "docs/*.md"
```

By default, README.md is set as the top page.
If needed, refer to the [configuration file specification](/yaml_specification) to modify the pages field.
Once your configuration file is ready, let's proceed to the final step of uploading.

# Uploading Documents

To upload documents, you need to set the API Key you obtained earlier as an environment variable named `DODO_API_KEY`.
Run the following command to set the API Key as an environment variable:

```bash
export DODO_API_KEY="<Your initially obtained API Key>"
```
:::message info
If you're uploading continuously from a local environment, using tools like (direnv)[https://direnv.net/] can be convenient.
:::

Now you're all set for uploading.
Let's run the following command to actually upload the documents:

```bash
dodo-cli upload
```

If successful, you'll see a log message saying `successfully uploaded`.
A URL for the document will also be displayed. Open this URL in your browser to check the result.

If you want to upload again, simply run `dodo-cli upload` once more.
Easy, isn't it?

# Next Steps
This concludes the basics of uploading.
If you want to learn more detailed usage, please refer to the links below.

https://document.do.dodo-doc.com/yaml_specification

https://document.do.dodo-doc.com/ci