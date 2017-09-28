# Copyright 2017 VMware, Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License

*** Settings ***
Documentation	Test 23-06 - VCH Certificate
Resource	../../resources/Util.robot
Resource	Util.robot
Suite Setup	Start VIC Machine Server
Suite Teardown	Terminate All Processes  kill=True
Default Tags


*** Keywords ***
Install VIC Machine Without TLS
    [Arguments]  ${vic-machine}=bin/vic-machine-linux  ${appliance-iso}=bin/appliance.iso  ${bootstrap-iso}=bin/bootstrap.iso  ${certs}=${true}  ${vol}=default  ${cleanup}=${true}  ${debug}=1  ${additional-args}=${EMPTY}
    Set Test Environment Variables
    Run Keyword If  '%{HOST_TYPE}' == 'ESXi'  Run  govc host.esxcli network firewall set -e false
    # Attempt to cleanup old/canceled tests
    Run Keyword If  ${cleanup}  Run Keyword And Ignore Error  Cleanup Dangling VMs On Test Server
    Run Keyword If  ${cleanup}  Run Keyword And Ignore Error  Cleanup Datastore On Test Server
    Run Keyword If  ${cleanup}  Run Keyword And Ignore Error  Cleanup Dangling Networks On Test Server
    Run Keyword If  ${cleanup}  Run Keyword And Ignore Error  Cleanup Dangling vSwitches On Test Server
    Run Keyword If  ${cleanup}  Run Keyword And Ignore Error  Cleanup Dangling Containers On Test Server
    Run Keyword If  ${cleanup}  Run Keyword And Ignore Error  Cleanup Dangling Resource Pools On Test Server
    Set Suite Variable  ${vicmachinetls}  --no-tls
    Log To Console  \nInstalling VCH to test server with tls disabled...
    ${output}=  Run VIC Machine Command  ${vic-machine}  ${appliance-iso}  ${bootstrap-iso}  ${certs}  ${vol}  ${debug}  ${additional-args}
    Log  ${output}
    Should Contain  ${output}  Installer completed successfully

    Get Docker Params  ${output}  ${certs}
    Log To Console  Installer completed successfully: %{VCH-NAME}...

    [Return]  ${output}


Get Certificate
    [Arguments]  ${vch-id}
    Get Path Under Target    vch/${vch-id}


Get Certificate Within Datacenter
    [Arguments]  ${vch-id}
    ${orig}=  Get Environment Variable  TEST_DATACENTER
    ${dc}=  Run Keyword If  '%{TEST_DATACENTER}' == '${SPACE}'  Get Datacenter Name
    Run Keyword If  '%{TEST_DATACENTER}' == '${SPACE}'  Set Environment Variable  TEST_DATACENTER  ${dc}
    Get Path Under Target    datacenter/%{TEST_DATACENTER}/vch/${vch-id}
    Set Environment Variable  TEST_DATACENTER  ${orig}


Verify Certificate
    Should Contain      ${OUTPUT}    BEGIN CERTIFICATE
    Should Contain      ${OUTPUT}    END CERTIFICATE
    Should Not Contain  ${OUTPUT}    "


*** Test Cases ***
Get VCH Certificate
    Install VIC Appliance To Test Server
    ${id}=  Get VCH ID  %{VCH-NAME}

    Get Certificate ${id}
    Verify Return Code
    Verify Status Ok
    Verify Certificate

    Get Certificate Within Datacenter ${id}
    Verify Return Code
    Verify Status Ok
    Verify Certificate

    Cleanup VIC Appliance On Test Server


Get VCH Certificate No TLS
    Install VIC Machine Without TLS
    ${id}=  Get VCH ID  %{VCH-NAME}

    Get Certificate ${id}
    Verify Return Code
    Verify Status 404

    Get Certificate Within Datacenter ${id}
    Verify Return Code
    Verify Status 404

    Cleanup VIC Appliance On Test Server
