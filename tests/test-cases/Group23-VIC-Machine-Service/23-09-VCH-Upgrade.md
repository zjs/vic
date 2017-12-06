Test 23-09 - VCH Upgrade
========================

#### Purpose:
To verify vic-machine-server properly handles upgrade and rollback of VCHs

#### References:
1. [VIC Machine Service API Design Doc](../../../doc/design/vic-machine/service.md)

#### Environment:
This suite requires a vSphere system where VCHs can be deployed.


Test Cases
----------------

###  1. Upgrade

#### Test Steps:
1. Create a VCH with an old version
2. Verify the expected version was deployed
3. Use the API to upgrade the VCH
4. Verify that the upgrade was successful
5. Verify that the VCH lists the new version

#### Expected Outcome:
* The VCH is successfully upgraded to the new version
