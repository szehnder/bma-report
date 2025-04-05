<script lang="ts">
	import { onMount } from 'svelte';

	interface Address {
		id: string;
		addressStr: string;
		enabled: boolean;
		primary: boolean;
	}

	interface BMAReport {
		primaryAddress: Address | null;
		comparisonAddresses: Address[] | null;
		opinion: string;
	}

	const API_URL = import.meta.env.VITE_API_URL;

	let addresses: Address[] = [];
	let bmaReport: BMAReport | null = null;
	let errorMessage: string = '';

	async function fetchAddresses() {
		try {
			const res = await fetch(`${API_URL}/api/addresses`);
			const data = await res.json();
			addresses = data;
		} catch (err) {
			errorMessage = 'Failed to fetch addresses.';
		}
	}

	async function updateAddress(id: string, updated: Partial<Address>) {
		try {
			console.log('Updating address:', { id, updated });
			const response = await fetch(`${API_URL}/api/addresses/${id}`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(updated)
			});
			console.log('Update response:', response);
			await fetchAddresses();
			await fetchBMAReport();
		} catch (err) {
			console.error('Error updating address:', err);
			errorMessage = 'Failed to update address.';
		}
	}

	async function fetchBMAReport() {
		try {
			const res = await fetch(`${API_URL}/api/bma-report`);
			bmaReport = await res.json();
		} catch (err) {
			errorMessage = 'Failed to fetch BMA report.';
		}
	}

	onMount(async () => {
		await fetchAddresses();
		await fetchBMAReport();
	});
</script>

<div>
	<h1>BMA Addresses</h1>
	{#if errorMessage}
		<p style="color: red">{errorMessage}</p>
	{/if}

	{#each addresses as addr}
		<div style="margin-bottom: 8px; border: 1px solid #ccc; padding: 8px">
			<p><strong>{addr.addressStr}</strong></p>
			<label>
				<input
					type="checkbox"
					checked={addr.enabled}
					on:change={() => updateAddress(addr.id, { enabled: !addr.enabled })}
				/>
				Enabled
			</label>
			<label style="margin-left: 10px;">
				<input
					type="radio"
					name="primary"
					checked={addr.primary}
					on:change={() => updateAddress(addr.id, { primary: true })}
				/>
				Set as Primary
			</label>
		</div>
	{/each}

	<h2>BMA Report</h2>
	{#if bmaReport}
		{#if bmaReport.primaryAddress && bmaReport.comparisonAddresses?.length > 0}
			<p><strong>Primary Address:</strong> {bmaReport.primaryAddress.addressStr}</p>
			<p><strong>Comparison Addresses:</strong></p>
			<ul>
				{#each bmaReport.comparisonAddresses as cAddr}
					<li>{cAddr.addressStr}</li>
				{/each}
			</ul>
			<div style="white-space: pre-line; margin-top: 10px;">
				{bmaReport.opinion}
			</div>
		{:else}
			<p>{bmaReport.opinion}</p>
		{/if}
	{/if}
</div>
