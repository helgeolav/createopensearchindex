# Create Opensearch indices mappings

This is a quick tool to create a mapping, either directly on an index or as a template. It can also
work as an HTTP server and create a mapping based on the JSON in the POST request.

The syntax for the input file is:

```json
{
  "input": {
    "field1": "type",
    "field2": "type",
    "nested1": {
      "nestedfield1": "type",
      "nestedfield2": "type"
    }
  }
}
```

Type is a supported Opensearch format.