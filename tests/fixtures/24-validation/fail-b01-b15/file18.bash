# Test B05 (field naming mismatch)
cp B05-field-naming-mismatch/python-client/models.py PERFECT/python-client/
pytest PERFECT/python-client/test_client.py -v
# Expected: ValidationError on stock_count field

# Test B08 (incomplete enum)
cp B08-incomplete-enum/python-client/models.py PERFECT/python-client/
pytest PERFECT/python-client/test_client.py -v
# Expected: ValidationError when backorder status returned

# Test B15 (missing ETag)
cp B15-missing-etag/python-client/client.py PERFECT/python-client/
pytest PERFECT/python-client/test_client.py::test_update_item_with_etag -v
# Expected: 428 Precondition Required error
