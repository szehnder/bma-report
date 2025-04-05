<script lang="ts">
	import { API_URL } from '../lib/constants';

	export let addresses: Array<{
		id: string;
		addressStr: string;
		enabled: boolean;
		primary: boolean;
		price?: number;
		bedrooms?: number;
		bathrooms?: number;
		squareFootage?: number;
		propertyType?: string;
		yearBuilt?: number;
	}>;
	export let onUpdate: () => Promise<void>;

	async function updateAddress(id: string, updates: { enabled?: boolean; primary?: boolean }) {
		try {
			const response = await fetch(`${API_URL}/api/addresses/${id}`, {
				method: 'PATCH',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify(updates),
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			await onUpdate();
		} catch (err) {
			console.error('Error updating address:', err);
		}
	}

	async function handleDelete(id: string) {
		if (!confirm('Are you sure you want to delete this address?')) {
			return;
		}

		try {
			const response = await fetch(`${API_URL}/api/addresses/${id}`, {
				method: 'DELETE',
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			await onUpdate();
		} catch (err) {
			console.error('Error deleting address:', err);
		}
	}

	async function handleSetPrimary(id: string) {
		const currentPrimary = addresses.find(a => a.primary);
		if (currentPrimary && currentPrimary.id !== id) {
			if (!confirm('Do you want to replace your current primary address with this one?')) {
				return;
			}
		}
		await updateAddress(id, { primary: true });
	}

	$: primaryAddress = addresses.find(a => a.primary);
	$: otherAddresses = addresses.filter(a => !a.primary);
</script>

<div class="address-list">
	{#if primaryAddress}
		<div class="address-group">
			<h3>Primary Address</h3>
			<div class="address-item primary">
				<div class="address-controls">
					<div class="controls-group">
						<span class="primary-label">Primary Address</span>
					</div>
					<button class="delete-button" on:click={() => handleDelete(primaryAddress.id)}>
						<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<path d="M3 6h18"></path>
							<path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path>
							<path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path>
						</svg>
					</button>
				</div>
				<div class="address-text">{primaryAddress.addressStr}</div>
				<div class="property-summary">
					<div class="summary-item">
						<span class="label">Price:</span>
						<span class="value">${primaryAddress.price?.toLocaleString() || 'N/A'}</span>
					</div>
					<div class="summary-item">
						<span class="label">Beds:</span>
						<span class="value">{primaryAddress.bedrooms || 'N/A'}</span>
					</div>
					<div class="summary-item">
						<span class="label">Baths:</span>
						<span class="value">{primaryAddress.bathrooms || 'N/A'}</span>
					</div>
					<div class="summary-item">
						<span class="label">Sq Ft:</span>
						<span class="value">{primaryAddress.squareFootage?.toLocaleString() || 'N/A'}</span>
					</div>
					<div class="summary-item">
						<span class="label">Type:</span>
						<span class="value">{primaryAddress.propertyType || 'N/A'}</span>
					</div>
					<div class="summary-item">
						<span class="label">Year:</span>
						<span class="value">{primaryAddress.yearBuilt || 'N/A'}</span>
					</div>
				</div>
			</div>
		</div>
	{/if}

	<div class="address-group">
		<h3>Comparison Addresses</h3>
		{#each otherAddresses as addr}
			<div class="address-item">
				<div class="address-controls">
					<div class="controls-group">
						<label class="checkbox-label">
							<input
								type="checkbox"
								checked={addr.enabled}
								on:change={() => updateAddress(addr.id, { enabled: !addr.enabled })}
							/>
							<span class="checkbox-custom"></span>
							Enabled
						</label>
						<button class="set-primary-button" on:click={() => handleSetPrimary(addr.id)}>
							Set as Primary
						</button>
					</div>
					<button class="delete-button" on:click={() => handleDelete(addr.id)}>
						<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<path d="M3 6h18"></path>
							<path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path>
							<path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path>
						</svg>
					</button>
				</div>
				<div class="address-text">{addr.addressStr}</div>
				<div class="property-summary">
					<div class="summary-item">
						<span class="label">Price:</span>
						<span class="value">${addr.price?.toLocaleString() || 'N/A'}</span>
					</div>
					<div class="summary-item">
						<span class="label">Beds:</span>
						<span class="value">{addr.bedrooms || 'N/A'}</span>
					</div>
					<div class="summary-item">
						<span class="label">Baths:</span>
						<span class="value">{addr.bathrooms || 'N/A'}</span>
					</div>
					<div class="summary-item">
						<span class="label">Sq Ft:</span>
						<span class="value">{addr.squareFootage?.toLocaleString() || 'N/A'}</span>
					</div>
					<div class="summary-item">
						<span class="label">Type:</span>
						<span class="value">{addr.propertyType || 'N/A'}</span>
					</div>
					<div class="summary-item">
						<span class="label">Year:</span>
						<span class="value">{addr.yearBuilt || 'N/A'}</span>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>

<style>
	.address-list {
		margin-top: 20px;
		font-family: 'Roboto', sans-serif;
	}

	.address-group {
		margin-bottom: 20px;
	}

	.address-group h3 {
		margin-bottom: 10px;
		color: #2c3e50;
		font-weight: 500;
		font-size: 1.2em;
		letter-spacing: 0.3px;
	}

	.address-item {
		margin-bottom: 8px;
		border: 1px solid #e0e0e0;
		padding: 12px;
		border-radius: 4px;
		position: relative;
		background: white;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
	}

	.address-item.primary {
		background-color: #f8f9fa;
		border-left: 4px solid #4CAF50;
	}

	.address-controls {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 8px;
	}

	.controls-group {
		display: flex;
		gap: 16px;
		align-items: center;
	}

	.checkbox-label {
		display: flex;
		align-items: center;
		gap: 4px;
		cursor: pointer;
		font-size: 0.9em;
		color: #495057;
	}

	.delete-button {
		background: none;
		border: none;
		padding: 4px;
		cursor: pointer;
		color: #dc3545;
		opacity: 0.7;
		transition: opacity 0.2s;
	}

	.delete-button:hover {
		opacity: 1;
	}

	.address-text {
		font-weight: 500;
		margin-bottom: 8px;
		color: #2c3e50;
		font-size: 1.1em;
	}

	.property-summary {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(100px, 1fr));
		gap: 8px;
		margin-top: 8px;
		padding: 12px;
		background: #f8f9fa;
		border-radius: 4px;
	}

	.summary-item {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.summary-item .label {
		font-size: 0.8em;
		color: #6c757d;
		font-weight: 400;
	}

	.summary-item .value {
		font-weight: 500;
		color: #495057;
	}

	.set-primary-button {
		background: none;
		border: 1px solid #4CAF50;
		color: #4CAF50;
		padding: 4px 8px;
		border-radius: 4px;
		cursor: pointer;
		font-size: 0.9em;
		transition: all 0.2s;
		font-weight: 500;
	}

	.set-primary-button:hover {
		background: #4CAF50;
		color: white;
	}

	.primary-label {
		color: #4CAF50;
		font-weight: 500;
		font-size: 0.9em;
	}
</style> 