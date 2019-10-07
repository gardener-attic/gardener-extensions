<p>Packages:</p>
<ul>
<li>
<a href="#certificate-service.extensions.config.gardener.cloud%2fv1alpha1">certificate-service.extensions.config.gardener.cloud/v1alpha1</a>
</li>
</ul>
<h2 id="certificate-service.extensions.config.gardener.cloud/v1alpha1">certificate-service.extensions.config.gardener.cloud/v1alpha1</h2>
<p>
<p>Package v1alpha1 contains the Certificate Service extension API resources.</p>
</p>
Resource Types:
<ul></ul>
<h3 id="certificate-service.extensions.config.gardener.cloud/v1alpha1.ACME">ACME
</h3>
<p>
(<em>Appears on:</em>
<a href="#certificate-service.extensions.config.gardener.cloud/v1alpha1.ConfigurationSpec">ConfigurationSpec</a>)
</p>
<p>
<p>ACME holds information about the ACME issuer used for the certificate service.</p>
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
<code>email</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>server</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>privateKey</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
<h3 id="certificate-service.extensions.config.gardener.cloud/v1alpha1.CloudDNS">CloudDNS
</h3>
<p>
(<em>Appears on:</em>
<a href="#certificate-service.extensions.config.gardener.cloud/v1alpha1.DNSProviders">DNSProviders</a>)
</p>
<p>
<p>CloudDNS is a DNS provider used for ACME DNS01 challenges.</p>
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
<code>domains</code></br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>project</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>serviceAccount</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="certificate-service.extensions.config.gardener.cloud/v1alpha1.Configuration">Configuration
</h3>
<p>
<p>Configuration contains information about the certificate service configuration.</p>
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
<code>spec</code></br>
<em>
<a href="#certificate-service.extensions.config.gardener.cloud/v1alpha1.ConfigurationSpec">
ConfigurationSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>lifecycleSync</code></br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>serviceSync</code></br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>issuerName</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>namespaceRef</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>resourceNamespace</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>acme</code></br>
<em>
<a href="#certificate-service.extensions.config.gardener.cloud/v1alpha1.ACME">
ACME
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>providers</code></br>
<em>
<a href="#certificate-service.extensions.config.gardener.cloud/v1alpha1.DNSProviders">
DNSProviders
</a>
</em>
</td>
<td>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="certificate-service.extensions.config.gardener.cloud/v1alpha1.ConfigurationSpec">ConfigurationSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#certificate-service.extensions.config.gardener.cloud/v1alpha1.Configuration">Configuration</a>)
</p>
<p>
<p>ConfigurationSpec contains information about the certificate service configuration.</p>
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
<code>lifecycleSync</code></br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>serviceSync</code></br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>issuerName</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>namespaceRef</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>resourceNamespace</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>acme</code></br>
<em>
<a href="#certificate-service.extensions.config.gardener.cloud/v1alpha1.ACME">
ACME
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>providers</code></br>
<em>
<a href="#certificate-service.extensions.config.gardener.cloud/v1alpha1.DNSProviders">
DNSProviders
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="certificate-service.extensions.config.gardener.cloud/v1alpha1.DNSProvider">DNSProvider
(<code>string</code> alias)</p></h3>
<p>
<p>DNSProvider string type</p>
</p>
<h3 id="certificate-service.extensions.config.gardener.cloud/v1alpha1.DNSProviderConfig">DNSProviderConfig
</h3>
<p>
<p>DNSProviderConfig is an interface that will implemented by cloud provider structs</p>
</p>
<h3 id="certificate-service.extensions.config.gardener.cloud/v1alpha1.DNSProviders">DNSProviders
</h3>
<p>
(<em>Appears on:</em>
<a href="#certificate-service.extensions.config.gardener.cloud/v1alpha1.ConfigurationSpec">ConfigurationSpec</a>)
</p>
<p>
<p>DNSProviders hold information about information about DNS providers used for ACME DNS01 challenges.</p>
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
<code>route53</code></br>
<em>
<a href="#certificate-service.extensions.config.gardener.cloud/v1alpha1.Route53">
[]Route53
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>cloudDNS</code></br>
<em>
<a href="#certificate-service.extensions.config.gardener.cloud/v1alpha1.CloudDNS">
[]CloudDNS
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="certificate-service.extensions.config.gardener.cloud/v1alpha1.Route53">Route53
</h3>
<p>
(<em>Appears on:</em>
<a href="#certificate-service.extensions.config.gardener.cloud/v1alpha1.DNSProviders">DNSProviders</a>)
</p>
<p>
<p>Route53 is a DNS provider used for ACME DNS01 challenges.</p>
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
<code>domains</code></br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>region</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>accessKeyID</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>secretAccessKey</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>9f5e77de</code>.
</em></p>
