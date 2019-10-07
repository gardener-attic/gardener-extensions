<p>Packages:</p>
<ul>
<li>
<a href="#calico.networking.extensions.gardener.cloud%2fv1alpha1">calico.networking.extensions.gardener.cloud/v1alpha1</a>
</li>
</ul>
<h2 id="calico.networking.extensions.gardener.cloud/v1alpha1">calico.networking.extensions.gardener.cloud/v1alpha1</h2>
<p>
<p>Package v1alpha1 contains the configuration of the Calico Network Extension.</p>
</p>
Resource Types:
<ul></ul>
<h3 id="calico.networking.extensions.gardener.cloud/v1alpha1.Backend">Backend
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.NetworkConfig">NetworkConfig</a>)
</p>
<p>
</p>
<h3 id="calico.networking.extensions.gardener.cloud/v1alpha1.CIDR">CIDR
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.IPAM">IPAM</a>)
</p>
<p>
</p>
<h3 id="calico.networking.extensions.gardener.cloud/v1alpha1.IPAM">IPAM
</h3>
<p>
(<em>Appears on:</em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.NetworkConfig">NetworkConfig</a>)
</p>
<p>
<p>IPAM defines the block that configuration for the ip assignment plugin to be used</p>
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
<code>type</code></br>
<em>
string
</em>
</td>
<td>
<p>Type defines the IPAM plugin type</p>
</td>
</tr>
<tr>
<td>
<code>cidr</code></br>
<em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.CIDR">
CIDR
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CIDR defines the CIDR block to be used</p>
</td>
</tr>
</tbody>
</table>
<h3 id="calico.networking.extensions.gardener.cloud/v1alpha1.NetworkConfig">NetworkConfig
</h3>
<p>
<p>NetworkConfig configuration for the calico networking plugin</p>
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
<code>backend</code></br>
<em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.Backend">
Backend
</a>
</em>
</td>
<td>
<p>Backend defines whether a backend should be used or not (e.g., bird or None)</p>
</td>
</tr>
<tr>
<td>
<code>ipam</code></br>
<em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.IPAM">
IPAM
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>IPAM to use for the Calico Plugin (e.g., host-local or Calico)</p>
</td>
</tr>
<tr>
<td>
<code>ipAutodetectionMethod</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>IPAutoDetectionMethod is the method to use to autodetect the IPv4 address for this host. This is only used when the IPv4 address is being autodetected.
<a href="https://docs.projectcalico.org/v2.2/reference/node/configuration#ip-autodetection-methods">https://docs.projectcalico.org/v2.2/reference/node/configuration#ip-autodetection-methods</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="calico.networking.extensions.gardener.cloud/v1alpha1.NetworkStatus">NetworkStatus
</h3>
<p>
<p>NetworkStatus contains information about created Network resources.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>9f5e77de</code>.
</em></p>
