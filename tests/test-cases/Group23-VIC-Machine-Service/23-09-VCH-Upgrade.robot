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
Documentation     Test 23-02 - VCH List
Resource          ../../resources/Util.robot
Resource          ../../resources/Group23-VIC-Machine-Service-Util.robot
Suite Setup       Start VIC Machine Server
Suite Teardown    Terminate All Processes    kill=True
Test Setup        Install VIC with version to Test Server    7315
Test Teardown     Cleanup VIC Appliance On Test Server
Default Tags


*** Keywords ***

Upgrade
    ${id}=    Get VCH ID %{VCH-NAME}

    Post Path Under Target    vch/${id}    ''    action="upgrade"


Verify Upgrade
    Verify Return Code
    Verify Status Accepted


*** Test Cases ***

Upgrade VCH
    Check Original Version

    Upgrade

    Verify Upgrade
    Check Upgraded Version
