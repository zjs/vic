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
Documentation	Test 23-01 - VCH Version
Resource	../../resources/Util.robot
Suite Setup	Start VIC Machine Server
Suite Teardown	Terminate All Processes  kill=True
Default Tags

*** Keywords ***
Start VIC Machine Server
    Start Process  ./bin/vic-machine-server --port 1337 --scheme http  shell=True  cwd=/go/src/github.com/vmware/vic
    Sleep  1s  for service to start

Curl Version
    [Arguments]
    ${rc}  ${output}=  Run And Return Rc And Output  curl -s -w "\n\%{http_code}\n" -X GET "http://127.0.0.1:1337/container/version"
    [Return]  ${rc}  ${output}

*** Test Cases ***
Get Version
    ${rc}  ${output}=  Curl Version
    Should Be Equal As Integers  ${rc}  0
    Should Match Regexp  ${output}  v\\d+\\.\\d+\\.\\d+-\\w+-\\d+-[a-f0-9]+
    Should Not Contain  ${output}  "
    ${status}=  Get Line  ${output}  -1
    Should Be Equal As Integers  200  ${status}
