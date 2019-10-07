<p>Packages:</p>
<ul>
<li>
<a href="#shoot-cert-service.extensions.config.gardener.cloud%2fv1alpha1">shoot-cert-service.extensions.config.gardener.cloud/v1alpha1</a>
</li>
</ul>
<h2 id="shoot-cert-service.extensions.config.gardener.cloud/v1alpha1">shoot-cert-service.extensions.config.gardener.cloud/v1alpha1</h2>
<p>
<p>Package v1alpha1 contains the Certificate Shoot Service extension configuration.</p>
</p>
Resource Types:
<ul></ul>
<h3 id="shoot-cert-service.extensions.config.gardener.cloud/v1alpha1.ACME">ACME
</h3>
<p>
(<em>Appears on:</em>
<a href="#shoot-cert-service.extensions.config.gardener.cloud/v1alpha1.Configuration">Configuration</a>)
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
<h3 id="shoot-cert-service.extensions.config.gardener.cloud/v1alpha1.Configuration">Configuration
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
<code>acme</code></br>
<em>
<a href="#shoot-cert-service.extensions.config.gardener.cloud/v1alpha1.ACME">
ACME
</a>
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
