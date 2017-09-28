Test 23-06 - VCH Certificate
=======

# Purpose:
To verify vic-machine-server can provide a VCH host certificate when available

# Test Steps:
1. Deploy the VCH into the test environment with tls enabled
2. Verify that the certificate is available using the vic-machine service
3. Verify that the certificate is available for its particular datacenter using the vic-machine service
4. Deploy a new VCH into the test environment using --no-tls and --no-tls-verify so that tls is disabled and no cert is created
5. Verify that the certificate is unavailable (404) using the vic-machine service
6. Verify that the certificate is unavailable (404) for its particular datacenter using the vic-machine service

# Expected Outcome:
* Step 2-3 should succeed and output should contain the host certificate
* Step 5-6 should error with 404 (not found) as no certs exist
