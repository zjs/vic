
oneGB = 1 * 1000 * 1000 # in KB

$testbed = Proc.new do |type, esxStyle, vcStyle, dbType, location|
  esxStyle ||= 'pxeBoot'
  vcStyle ||= 'vpxInstall'

  esxStyle = esxStyle.to_s
  vcStyle = vcStyle.to_s
  sharedStorageStyle = 'iscsi'
  dbType = dbType.to_s

  nameParts = ['vic', 'vsan', type, esxStyle, vcStyle]
  if dbType != 'embedded'
    nameParts << dbType
  end

  testbed = {
    "name" => nameParts.join('-'),
    "version" => 3,
    "esx" => (0..3).map do | idx |
      {
        "name" => "esx.#{idx}",
        "vc" => "vc.0",
        "style" => "pxeInstall",
        "desiredPassword" => "e2eFunctionalTest",
        'numMem' => 13 * 1024,
        'numCPUs' => 4,
        "disks" => [ 30 * oneGB, 30 * oneGB, 30 * oneGB],
        'freeLocalLuns' => 1,
        'freeSharedLuns' => 2,
        'ssds' => [ 5*oneGB ],
        'nics' => 2,
        'vmotionNics' => ['vmk0'],
        'staf' => false,
        "mountNfs" => ["nfs.0"],
        "clusterName" => "cls",
      }
    end,

    "nfs" => [
      {
        "name" => "nfs.0",
        "type" => "NFS41"
      }
    ],

    "vcs" => [
      {
        "name" => "vc.0",
        "type" => "vcva",
        "dcName" => "dc1",
        "clusters" => [{"name" => "cls", "vsan" => true, "enableDrs" => true, "enableHA" => true}],
        "addHosts" => "allInSameCluster",
      }
    ],

    'vsan' => true,
    'vsanEnable' => true, # pulled from test-vpx/testbeds/vc-on-vsan.rb

    "postBoot" => Proc.new do |runId, testbedSpec, vmList, catApi, logDir|
      esxList = vmList['esx']
        esxList.each do |host|
          host.ssh do |ssh|
            ssh.exec!("esxcli network firewall set -e false")
          end
      end
      vc = vmList['vc'][0]
      vim = VIM.connect vc.rbvmomiConnectSpec
      datacenters = vim.serviceInstance.content.rootFolder.childEntity.grep(RbVmomi::VIM::Datacenter)
      raise "Couldn't find a Datacenter precreated"  if datacenters.length == 0
      datacenter = datacenters.first
      Log.info "Found a datacenter successfully in the system, name: #{datacenter.name}"
      clusters = datacenter.hostFolder.children
      raise "Couldn't find a cluster precreated"  if clusters.length == 0
      cluster = clusters.first
      Log.info "Found a cluster successfully in the system, name: #{cluster.name}"

      dvs = datacenter.networkFolder.CreateDVS_Task(
        :spec => {
          :configSpec => {
            :name => "test-ds"
          },
            }
      ).wait_for_completion
      Log.info "Vds DSwitch created"

      dvpg1 = dvs.AddDVPortgroup_Task(
        :spec => [
          {
            :name => "management",
            :type => :earlyBinding,
            :numPorts => 12,
          }
        ]
      ).wait_for_completion
      Log.info "management DPG created"

      dvpg2 = dvs.AddDVPortgroup_Task(
        :spec => [
          {
            :name => "vm-network",
            :type => :earlyBinding,
            :numPorts => 12,
          }
        ]
      ).wait_for_completion
      Log.info "vm-network DPG created"

      dvpg3 = dvs.AddDVPortgroup_Task(
        :spec => [
          {
            :name => "bridge",
            :type => :earlyBinding,
            :numPorts => 12,
          }
        ]
      ).wait_for_completion
      Log.info "bridge DPG created"

      Log.info "Add hosts to the DVS"
      onecluster_pnic_spec = [ VIM::DistributedVirtualSwitchHostMemberPnicSpec({:pnicDevice => 'vmnic1'}) ]
      dvs_config = VIM::DVSConfigSpec({
        :configVersion => dvs.config.configVersion,
        :host => cluster.host.map do |host|
        {
          :operation => :add,
          :host => host,
          :backing => VIM::DistributedVirtualSwitchHostMemberPnicBacking({
            :pnicSpec => onecluster_pnic_spec
          })
        }
        end
      })
      dvs.ReconfigureDvs_Task(:spec => dvs_config).wait_for_completion
      Log.info "Hosts added to DVS successfully"
    end
  }

  testbed
end

