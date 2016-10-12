*** Settings ***
Documentation  Test 13-1 - VIC UI Installation
Resource  ../../resources/Util.robot
Resource  ./vicui-common.robot
Test Teardown  Cleanup Installer Environment
#Suite Setup  Install VIC Appliance To Test Server
#Suite Teardown  Cleanup VIC Appliance On Test Server

*** Test Cases ***
Check Configs
    # Store the original configs file content in a variable
    # Set the exact paths to the installer / uninstaller scripts for use with tests
    Run Keyword  Set Absolute Script Paths

Ensure Vicui Plugin Is Not Registered Before Testing
    ${configs_c}=  Set Vcenter Ip
    Log To Console  \nwho am i\n=============================
    ${log_whoami}=  Run  whoami
    Log To Console  ${log_whoami}
    Log To Console  \nrunning chmod -R 0777 ${UI_INSTALLER_PATH}\n=============================
    Run  chmod -R 0777 ${UI_INSTALLER_PATH}
    Log To Console  \npermissions check\n=============================
    ${log0}=  Run  ls -la ${UI_INSTALLER_PATH}/..
    ${log1}=  Run  ls -la ${UI_INSTALLER_PATH}
    Log To Console  ${log0}
    Log To Console  ${log1}
    Log To Console  \nwhich shell\n=============================
    ${log2}=  Run  ps -fp $$
    Log To Console  ${log2}
    Log To Console  \ntry writing configs data\n=============================
    ${rm_configs}=  Run  rm -v ${UI_INSTALLER_PATH}/configs
    Log To Console  ${rm_configs}
    ${rc}=  Run And Return Rc  echo '${configs_c}' | tee ${UI_INSTALLER_PATH}/configs
    #${status}=  Run Keyword And Return Status  Create File  ${UI_INSTALLER_PATH}/configs  ${configs_c}
    ${d}=  OperatingSystem.Get File  ${UI_INSTALLER_PATH}/configs
    Log To Console  ${d}
    Run Keyword  Force Remove Vicui Plugin

Attempt To Install With Configs File Missing
    # Rename the configs file and run the installer script to see if it fails in an expected way
    Move File  ${UI_INSTALLER_PATH}/configs  ${UI_INSTALLER_PATH}/configs_renamed
    ${rc}  ${output}=  Run And Return Rc And Output  ${INSTALLER_SCRIPT_PATH}
    Run Keyword And Continue On Failure  Should Contain  ${output}  Configs file is missing
    Move File  ${UI_INSTALLER_PATH}/configs_renamed  ${UI_INSTALLER_PATH}/configs

Attempt To Install With Plugin Missing
    # Rename the folder containing the VIC UI binaries and run the installer script to see if it fails in an expected way
    #Set Vcenter Ip
    Move Directory  ${UI_INSTALLER_PATH}/../vsphere-client-serenity  ${UI_INSTALLER_PATH}/../vsphere-client-serenity-a
    Install Fails At Extension Reg  ${TEST_VC_USERNAME}  ${TEST_VC_PASSWORD}  ${TEST_VC_ROOT_PASSWORD}  ${TRUE}
    ${output}=  OperatingSystem.GetFile  install.log
    Run Keyword And Continue On Failure  Should Contain  ${output}  VIC UI plugin bundle was not found
    Move Directory  ${UI_INSTALLER_PATH}/../vsphere-client-serenity-a  ${UI_INSTALLER_PATH}/../vsphere-client-serenity
    Remove File  install.log

Attempt To Install With vCenter IP Missing
    # Leave VCENTER_IP empty and run the installer script to see if it fails in an expected way
    Remove File  ${UI_INSTALLER_PATH}/configs
    ${results}=  Replace String Using Regexp  ${configs}  VCENTER_IP=.*  VCENTER_IP=\"\"
    Create File  ${UI_INSTALLER_PATH}/configs  ${results}
    ${rc}  ${output}=  Run And Return Rc And Output  cd ${UI_INSTALLER_PATH} && ${INSTALLER_SCRIPT_PATH}
    Run Keyword And Continue On Failure  Should Contain  ${output}  Please provide a valid IP

Attempt To Install With Invalid vCenter IP
    # Populate VCENTER_IP with an invalid hostname and run the installer script to see if it fails in an expected way
    Remove File  ${UI_INSTALLER_PATH}/configs
    ${results}=  Replace String Using Regexp  ${configs}  VCENTER_IP=.*  VCENTER_IP=\"i-am-not-a-valid-ip\"
    Create File  ${UI_INSTALLER_PATH}/configs  ${results}
    Install Fails For Wrong Vcenter Ip  ${TEST_VC_USERNAME}  ${TEST_VC_PASSWORD}  ${TEST_VC_ROOT_PASSWORD}
    ${output}=  OperatingSystem.GetFile  install.log
    Should Contain  ${output}  Could not resolve the hostname
    Remove File  install.log

Attempt To Install With Wrong Vcenter Credentials
    # Try installing the plugin with wrong vCenter credentials and see if it fails in an expected way
    Remove File  ${UI_INSTALLER_PATH}/configs
    ${results}=  Replace String Using Regexp  ${configs}  VCENTER_IP=.*  VCENTER_IP=\"${TEST_VC_IP}\"
    Create File  ${UI_INSTALLER_PATH}/configs  ${results}
    #Set Vcenter Ip
    Install Fails At Extension Reg  ${TEST_VC_USERNAME}_nope  ${TEST_VC_PASSWORD}_nope  ${TEST_VC_ROOT_PASSWORD}  ${TRUE}
    ${output}=  OperatingSystem.GetFile  install.log
    Should Contain  ${output}  Cannot complete login due to an incorrect user name or password
    Remove File  install.log

Attempt To Install With Wrong Root Password
    # Try installing the plugin with wrong vCenter root password and see if it fails in an expected way
    Log To Console  Skipping this test, as making three incorrect attempts will lock the root account for a certain amount of time
    #Set Vcenter Ip
    #Install Vicui Without Webserver  ${TEST_VC_USERNAME}  ${TEST_VC_PASSWORD}  ${TEST_VC_ROOT_PASSWORD}_abc
    #${output}=  OperatingSystem.GetFile  install.log
    #Should Contain  ${output}  Root password is incorrect
    #Remove File  install.log

Attempt To Install Without Webserver Nor Bash Support
    # Try installing the plugin against a VCSA that has Bash disabled and see if it fails gracefully with instructions
    [Timeout]  ${TIMEOUT}
    #Set Vcenter Ip
    Append To File  ${UI_INSTALLER_PATH}/configs  SIMULATE_NO_BASH_SUPPORT=1\n
    Install Vicui Without Webserver Nor Bash  ${TEST_VC_USERNAME}  ${TEST_VC_PASSWORD}  ${TEST_VC_ROOT_PASSWORD}
    ${output}=  OperatingSystem.GetFile  install.log
    Run Keyword And Continue On Failure  Should Contain  ${output}  Bash shell is required
    Force Remove Vicui Plugin
    Remove File  install.log

Install Successfully Without Webserver
    [Timeout]  ${TIMEOUT}
    Remove File  ${UI_INSTALLER_PATH}/configs
    ${results}=  Replace String Using Regexp  ${configs}  VCENTER_IP=.*  VCENTER_IP=\"${TEST_VC_IP}\"
    Create File  ${UI_INSTALLER_PATH}/configs  ${results}
    #Set Vcenter Ip
    Install Vicui Without Webserver  ${TEST_VC_USERNAME}  ${TEST_VC_PASSWORD}  ${TEST_VC_ROOT_PASSWORD}
    ${output}=  OperatingSystem.GetFile  install.log
    ${run_results}=  Run Keyword And Return Status  Should Contain  ${output}  was successful
    ${now}=  Run  date +%m%d%y-%H%M%S
    Run Keyword If  ${run_results} == False  Copy File  install.log  install-FAIL-${now}.log
    Remove File  install.log

Attempt To Install When Plugin Is Already Registered
    # Plugin is already installed at this moment on he target VCSA
    # Try installing the plugin and see if it fails in an expected way
    [Timeout]  ${TIMEOUT}
    #Set Vcenter Ip
    Install Fails At Extension Reg  ${TEST_VC_USERNAME}  ${TEST_VC_PASSWORD}  ${TEST_VC_ROOT_PASSWORD}  ${TRUE}
    ${output}=  OperatingSystem.GetFile  install.log
    Should Contain  ${output}  is already registered
    Remove File  install.log

Install Successfully Without Webserver Using Force Flag
    # Plugin is already installed at this moment on he target VCSA
    # Install the plugin using the --force flag and see if it succeeds
    [Timeout]  ${TIMEOUT}
    #Set Vcenter Ip
    Install Vicui Without Webserver  ${TEST_VC_USERNAME}  ${TEST_VC_PASSWORD}  ${TEST_VC_ROOT_PASSWORD}  ${TRUE}
    ${output}=  OperatingSystem.GetFile  install.log
    ${run_results}=  Run Keyword And Return Status  Should Contain  ${output}  was successful
    ${now}=  Run  date +%m%d%y-%H%M%S
    Run Keyword If  ${run_results} == False  Copy File  install.log  install-FAIL-${now}.log
    Remove File  install.log
    Log To Console  Force removing Vicui for next tests...
    Force Remove Vicui Plugin

Attempt To Install With Webserver And Wrong Path To Plugin
    # Try installing the plugin using a web server while providing VIC_UI_HOST_URL that does not exist and see if it fails in an expected way
    #Set Vcenter Ip
    Remove File  ${UI_INSTALLER_PATH}/configs
    ${results}=  Replace String Using Regexp  ${configs}  VCENTER_IP=.*  VCENTER_IP=\"${TEST_VC_IP}\"
    ${results}=  Replace String Using Regexp  ${results}  VIC_UI_HOST_URL=.*  VIC_UI_HOST_URL=\"http:\/\/this-fake-host\.does-not-exist\"
    Create File  ${UI_INSTALLER_PATH}/configs  ${results}
    Wait Until Created  ${UI_INSTALLER_PATH}/configs
    Install Fails At Extension Reg  ${TEST_VC_USERNAME}  ${TEST_VC_PASSWORD}  ${TEST_VC_ROOT_PASSWORD}  ${FALSE}
    ${output}=  OperatingSystem.GetFile  install.log
    Should Contain  ${output}  Could not resolve the host
    Remove File  install.log

