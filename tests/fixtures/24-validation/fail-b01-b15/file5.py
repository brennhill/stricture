# Python sends:
{"stock_count": 150.0}

# Go receives and stores as int64(150)
# On next read, Python gets:
{"stock_count": 150}

# Can cause validation issues if constraints differ
