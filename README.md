# Create Opensearch indices mappings

This is a quick tool to create a mapping, either directly on an index or as a template. It can also
work as an HTTP server and create a mapping based on the JSON in the PUT request.

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

You create mappings as follows:
1. When creating a mapping directly for an index, do a PUT /index-name with the output in the body.
1. Index templates create with PUT /_index_template/your_template_name and after that when new indices are created matching "index_patters" the mappings are added automatically.