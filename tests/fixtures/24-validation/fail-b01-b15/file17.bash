cd PERFECT/python-client

# Start Go server first
cd ../go-server
go run . &
SERVER_PID=$!

# Run Python tests
cd ../python-client
pytest test_client.py -v

# Stop server
kill $SERVER_PID
