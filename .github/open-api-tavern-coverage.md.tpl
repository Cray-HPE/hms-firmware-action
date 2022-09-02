<!-- This file is templated with https://pkg.go.dev/html/template -->

# Open-API Tavern Coverage Report
<table>
	<tbody>
		<tr>
			<td>Endpoint</td>
			<td>Method</td>
			<td>Test Case Count</td>
		</tr>
{{- range $endpoint := .endpoints }}
    <tr>
        <td>{{$endpoint.url}}</td>
        <td>{{$endpoint.method}}</td>
        <td>{{$endpoint.count}}</td>
    </tr>
{{- end}}
	</tbody>
</table>