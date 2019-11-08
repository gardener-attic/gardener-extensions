<p>Packages:</p>
<ul>
<li>
<a href="#alicloud.provider.extensions.config.gardener.cloud%2fv1alpha1">alicloud.provider.extensions.config.gardener.cloud/v1alpha1</a>
</li>
</ul>
<h2 id="alicloud.provider.extensions.config.gardener.cloud/v1alpha1">alicloud.provider.extensions.config.gardener.cloud/v1alpha1</h2>
<p>
<p>Package v1alpha1 contains the AliCloud provider configuration API resources.</p>
</p>
Resource Types:
<ul><li>
<a href="#alicloud.provider.extensions.config.gardener.cloud/v1alpha1.ControllerConfiguration">ControllerConfiguration</a>
</li></ul>
<h3 id="alicloud.provider.extensions.config.gardener.cloud/v1alpha1.ControllerConfiguration">ControllerConfiguration
</h3>
<p>
<p>ControllerConfiguration defines the configuration for the Alicloud provider.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
alicloud.provider.extensions.config.gardener.cloud/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>ControllerConfiguration</code></td>
</tr>
<tr>
<td>
<code>clientConnection</code></br>
<em>
<a href="https://godoc.org/k8s.io/component-base/config/v1alpha1#ClientConnectionConfiguration">
Kubernetes v1alpha1.ClientConnectionConfiguration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ClientConnection specifies the kubeconfig file and client connection
settings for the proxy server to use when communicating with the apiserver.</p>
</td>
</tr>
<tr>
<td>
<code>machineImages</code></br>
<em>
<a href="#alicloud.provider.extensions.config.gardener.cloud/v1alpha1.MachineImage">
[]MachineImage
</a>
</em>
</td>
<td>
<p>MachineImages is the list of machine images that are understood by the controller. It maps
logical names and versions to Alicloud-specific identifiers.</p>
</td>
</tr>
<tr>
<td>
<code>machineImageOwnerSecretRef</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#secretreference-v1-core">
Kubernetes core/v1.SecretReference
</a>
</em>
</td>
<td>
<p>MachineImageOwnerSecretRef is the secret reference which contains credential of AliCloud subaccount for customized images.
We currently assume multiple customized images should always be under this account.</p>
</td>
</tr>
<tr>
<td>
<code>etcd</code></br>
<em>
<a href="#alicloud.provider.extensions.config.gardener.cloud/v1alpha1.ETCD">
ETCD
</a>
</em>
</td>
<td>
<p>ETCD is the etcd configuration.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="alicloud.provider.extensions.config.gardener.cloud/v1alpha1.ETCD">ETCD
</h3>
<p>
(<em>Appears on:</em>
<a href="#alicloud.provider.extensions.config.gardener.cloud/v1alpha1.ControllerConfiguration">ControllerConfiguration</a>)
</p>
<p>
<p>ETCD is an etcd configuration.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>storage</code></br>
<em>
<a href="#alicloud.provider.extensions.config.gardener.cloud/v1alpha1.ETCDStorage">
ETCDStorage
</a>
</em>
</td>
<td>
<p>ETCDStorage is the etcd storage configuration.</p>
</td>
</tr>
<tr>
<td>
<code>backup</code></br>
<em>
<a href="#alicloud.provider.extensions.config.gardener.cloud/v1alpha1.ETCDBackup">
ETCDBackup
</a>
</em>
</td>
<td>
<p>ETCDBackup is the etcd backup configuration.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="alicloud.provider.extensions.config.gardener.cloud/v1alpha1.ETCDBackup">ETCDBackup
</h3>
<p>
(<em>Appears on:</em>
<a href="#alicloud.provider.extensions.config.gardener.cloud/v1alpha1.ETCD">ETCD</a>)
</p>
<p>
<p>ETCDBackup is an etcd backup configuration.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>schedule</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Schedule is the etcd backup schedule.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="alicloud.provider.extensions.config.gardener.cloud/v1alpha1.ETCDStorage">ETCDStorage
</h3>
<p>
(<em>Appears on:</em>
<a href="#alicloud.provider.extensions.config.gardener.cloud/v1alpha1.ETCD">ETCD</a>)
</p>
<p>
<p>ETCDStorage is an etcd storage configuration.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>className</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ClassName is the name of the storage class used in etcd-main volume claims.</p>
</td>
</tr>
<tr>
<td>
<code>capacity</code></br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/api/resource#Quantity">
k8s.io/apimachinery/pkg/api/resource.Quantity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Capacity is the storage capacity used in etcd-main volume claims.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="alicloud.provider.extensions.config.gardener.cloud/v1alpha1.MachineImage">MachineImage
</h3>
<p>
(<em>Appears on:</em>
<a href="#alicloud.provider.extensions.config.gardener.cloud/v1alpha1.ControllerConfiguration">ControllerConfiguration</a>)
</p>
<p>
<p>MachineImage is a mapping from logical names and versions to Alicloud-specific identifiers.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name is the logical name of the machine image.</p>
</td>
</tr>
<tr>
<td>
<code>version</code></br>
<em>
string
</em>
</td>
<td>
<p>Version is the logical version of the machine image.</p>
</td>
</tr>
<tr>
<td>
<code>regions</code></br>
<em>
<a href="#alicloud.provider.extensions.config.gardener.cloud/v1alpha1.RegionImageMapping">
[]RegionImageMapping
</a>
</em>
</td>
<td>
<p>Regions is a mapping to the correct IDs for the machine image in the supported regions.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="alicloud.provider.extensions.config.gardener.cloud/v1alpha1.RegionImageMapping">RegionImageMapping
</h3>
<p>
(<em>Appears on:</em>
<a href="#alicloud.provider.extensions.config.gardener.cloud/v1alpha1.MachineImage">MachineImage</a>)
</p>
<p>
<p>RegionImageMapping is a mapping from Region name to supported machine image id for a specific OS version.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>region</code></br>
<em>
string
</em>
</td>
<td>
<p>Region is the ID of the region.</p>
</td>
</tr>
<tr>
<td>
<code>imageID</code></br>
<em>
string
</em>
</td>
<td>
<p>ImageID is the ID for the machine image.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
