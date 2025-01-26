# Formatting

*Heimdall* uses JSON as its default output format (unless defined otherwise by a command), but offers other formats.

Use the `--output` (or `-o`) parameter to format CLI output.
The argument values and types of output are:

* `csv` - Comma-separated key-value pairs.
* `json` - JSON string. This setting is the default.
* `table` - Presents the information in a "human-friendly" format that is much easier to read than the others, but not as programmatically useful.
* `template` - Format the output using a Go template expression.
* `template-file` - Format the output using a Go template from a file or URL.
* `text` - Plain text output without special formatting.
* `tsv` - Tab-separated key-value pairs (useful for `grep`, `sed`, or `awk`).
* `yaml` - YAML, a machine-readable alternative to JSON.

### Filter Output

The *Heimdall* CLI has two built-in JSON-based client-side filtering capabilities.
The optional parameters are a powerful tool you can use to customize the content and style of your output.
They take the JSON representation of the internal command-specific data structures and filter the results before displaying them.

* The `--query` parameter uses [JMESPath](http://jmespath.org/) syntax to create expressions for filtering your output.
  To learn JMESPath syntax, see [Tutorial](https://jmespath.org/tutorial.html) on the JMESPath website.
* The `--jq` parameter uses [JQ](https://jqlang.github.io/jq/) syntax for general purpose JSON processing.
  The [JQ manual](https://jqlang.github.io/jq/manual/) provides a comprehensive overview.

The following code listing show examples of what the `--query` and `--jq` parameter can produce.

```shell
$ env | grep -E "^USER=" | heimdall properties --output json
{"USER":"me"}
$ env | heimdall properties --output text --query "USER"
me
$ env | heimdall properties --output text --jq ".USER"
me
```
