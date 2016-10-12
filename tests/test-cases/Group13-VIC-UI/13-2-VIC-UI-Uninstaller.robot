*** Settings ***
Documentation  Test 13-2 - VIC UI Uninstallation
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

Ensure Vicui Is Installed Before Testing
    Set Vcenter Ip
    Install Vicui Without Webserver  ${TEST_VC_USERNAME}  ${TEST_VC_PASSWORD}  ${TEST_VC_ROOT_PASSWORD}  ${TRUE}
    ${output}=  OperatingSystem.GetFile  install.log
    Should Contain  ${output}  was successful
    Remove File  install.log
    
Attempt To Uninstall With Configs File Missing
    # Rename the configs file and run the uninstaller script to see if it fails in an expected way
    Move File  ${UI_INSTALLER_PATH}/configs  ${UI_INSTALLER_PATH}/configs_renamed
    ${rc}  ${output}=  Run And Return Rc And Output  ${UNINSTALLER_SCRIPT_PATH}
    Run Keyword And Continue On Failure  Should Contain  ${output}  Configs file is missing
    Move File  ${UI_INSTALLER_PATH}/configs_renamed  ${UI_INSTALLER_PATH}/configs

Attempt To Uninstall With Plugin Missing
    # Rename the folder containing the VIC UI binaries and run the uninstaller script to see if it fails in an expected way
    Set Vcenter Ip
    Move Directory  ${UI_INSTALLER_PATH}/../vsphere-client-serenity  ${UI_INSTALLER_PATH}/../vsphere-client-serenity-a
    Uninstall Fails  ${TEST_VC_USERNAME}  ${TEST_VC_PASSWORD}
    ${output}=  OperatingSystem.GetFile  uninstall.log
    Run Keyword And Continue On Failure  Should Contain  ${output}  VIC UI plugin bundle was not found
    Move Directory  ${UI_INSTALLER_PATH}/../vsphere-client-serenity-a  ${UI_INSTALLER_PATH}/../vsphere-client-serenity
    Remove File  uninstall.log

Attempt To Uninstall With vCenter IP Missing
    # Leave VCENTER_IP empty and run the uninstaller script to see if it fails in an expected way
    Remove File  ${UI_INSTALLER_PATH}/configs
    ${results}=  Replace String Using Regexp  ${configs}  VCENTER_IP=.*  VCENTER_IP=\"\"
    Create File  ${UI_INSTALLER_PATH}/configs  ${results}
    ${rc}  ${output}=  Run And Return Rc And Output  cd ${UI_INSTALLER_PATH} && ${UNINSTALLER_SCRIPT_PATH}
    Run Keyword And Continue On Failure  Should Contain  ${output}  Please provide a valid IP

Attempt To Uninstall With Wrong Vcenter Credentials
    # Try uninstalling the plugin with wrong vCenter credentials and see if it fails in an expected way
    Set Vcenter Ip
    Uninstall Fails  ${TEST_VC_USERNAME}_nope  ${TEST_VC_PASSWORD}_nope
    ${output}=  OperatingSystem.GetFile  uninstall.log
    Should Contain  ${output}  Cannot complete login due to an incorrect user name or password
    Remove File  uninstall.log

Uninstall Successfully
    Set Vcenter Ip
    Uninstall Vicui  ${TEST_VC_USERNAME}  ${TEST_VC_PASSWORD}
    ${output}=  OperatingSystem.GetFile  uninstall.log
    Should Match Regexp  ${output}  unregistration was successful
    Remove File  uninstall.log

Attempt To Uninstall Plugin That Is Already Gone
    Set Vcenter Ip
    Uninstall Fails  ${TEST_VC_USERNAME}  ${TEST_VC_PASSWORD}
    ${output}=  OperatingSystem.GetFile  uninstall.log
    Should Contain  ${output}  failed to find target
