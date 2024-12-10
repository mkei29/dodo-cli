
On this page, we’ll show you how to automate uploading your documentation using GitHub Actions.

# Step1. Prepare the workflow file
Start by creating a YAML file for your CI process under `.github/workflows`.
In the following example, we'll use `.github/workflows/publish-docs.yaml`.

To automatically upload your documentation, include the YAML content shown below.
In this example, the documentation will be uploaded automatically whenever a commit is pushed to the main branch.

```yaml
name: publish-docs
on:
  push:
    branches:
      - main
jobs:
  publish-relesae-notes:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v2
      - name: Install dodo CLI
        run: |
          curl https://raw.githubusercontent.com/toritoritori29/dodo-cli/main/download.sh | sh -
      - name: Publish docs
        run: |
          ./dodo-cli upload
        env:
          DODO_API_KEY: ${{ secrets.DODO_API_KEY }}
```

# Step2. Publish API Key and register to github

Follow the documentation to issue an API key, which will start with the prefix `ds-`.

Once issued, register the API key as a GitHub Actions secret so that it can be accessed by the workflow.
Following the instructions in the documentation, set the name of the secret to `DODO_API_KEY`.
