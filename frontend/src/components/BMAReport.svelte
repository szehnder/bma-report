<script lang="ts">
	import { API_URL } from '../lib/constants';

	export let bmaReport: {
		summary: string;
		detailedAnalysis: {
			primaryPropertyDetails: {
				address: string;
				price: number;
				bedrooms: number;
				bathrooms: number;
				squareFootage: number;
				yearBuilt: number;
				propertyType: string;
				lotSize: string;
				mlsNumber: string;
				daysOnMarket: number;
				lastPriceChange: number;
				description: string;
			};
			comparisonDetails: Array<{
				address: string;
				price: number;
				bedrooms: number;
				bathrooms: number;
				squareFootage: number;
				yearBuilt: number;
				propertyType: string;
				lotSize: string;
				mlsNumber: string;
				daysOnMarket: number;
				lastPriceChange: number;
				description: string;
			}>;
			priceAnalysis: string;
			featureComparison: Array<{
				feature: string;
				primaryValue: string;
				comparison: Array<{
					address: string;
					value: string;
				}>;
				analysis: string;
			}>;
			marketTrends: string;
			recommendation: string;
		};
	} | null = null;
	export let isLoadingBMA = false;
	export let errorMessage = '';

	async function refreshReport() {
		if (isLoadingBMA) return;
		
		isLoadingBMA = true;
		errorMessage = '';
		
		try {
			const response = await fetch(`${API_URL}/api/bma-report/refresh`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				}
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			const data = await response.json();
			bmaReport = data;
		} catch (err) {
			console.error('Error refreshing report:', err);
			errorMessage = 'Failed to refresh report. Please try again.';
		} finally {
			isLoadingBMA = false;
		}
	}

	function handleDownload() {
		if (!bmaReport) return;
		
		// Create a new window for printing
		const printWindow = window.open('', '_blank');
		if (!printWindow) return;

		// Write the HTML content
		printWindow.document.write(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>BMA Report</title>
				<style>
					@import url('https://fonts.googleapis.com/css2?family=Roboto:wght@300;400;500;700&display=swap');
					body {
						font-family: 'Roboto', sans-serif;
						line-height: 1.6;
						padding: 20px;
						max-width: 800px;
						margin: 0 auto;
					}
					h1, h2, h3 {
						color: #2c3e50;
					}
					h1 {
						font-size: 24px;
						margin-bottom: 20px;
					}
					h2 {
						font-size: 20px;
						margin-top: 20px;
						margin-bottom: 10px;
					}
					.property-details {
						background: #f8f9fa;
						padding: 15px;
						border-radius: 4px;
						margin-bottom: 20px;
					}
					.comparison-table {
						width: 100%;
						border-collapse: collapse;
						margin: 20px 0;
					}
					.comparison-table th, .comparison-table td {
						padding: 8px;
						border: 1px solid #e0e0e0;
						text-align: left;
					}
					.comparison-table th {
						background: #f8f9fa;
					}
					.recommendation {
						background: #e8f5e9;
						padding: 15px;
						border-radius: 4px;
						margin-top: 20px;
					}
					@media print {
						body {
							padding: 0;
						}
						@page {
							margin: 1cm;
						}
					}
				</style>
			</head>
			<body>
				<h1>Broker Market Analysis Report</h1>
				
				<h2>Primary Property</h2>
				<div class="property-details">
					<p><strong>Address:</strong> ${bmaReport.detailedAnalysis.primaryPropertyDetails.address}</p>
					<p><strong>Price:</strong> $${bmaReport.detailedAnalysis.primaryPropertyDetails.price.toLocaleString()}</p>
					<p><strong>Details:</strong> ${bmaReport.detailedAnalysis.primaryPropertyDetails.bedrooms} beds, 
						${bmaReport.detailedAnalysis.primaryPropertyDetails.bathrooms} baths, 
						${bmaReport.detailedAnalysis.primaryPropertyDetails.squareFootage.toLocaleString()} sq ft</p>
					<p><strong>Property Type:</strong> ${bmaReport.detailedAnalysis.primaryPropertyDetails.propertyType}</p>
					<p><strong>Year Built:</strong> ${bmaReport.detailedAnalysis.primaryPropertyDetails.yearBuilt}</p>
				</div>

				<h2>Price Analysis</h2>
				<p>${bmaReport.detailedAnalysis.priceAnalysis}</p>

				<h2>Feature Comparison</h2>
				<table class="comparison-table">
					<thead>
						<tr>
							<th>Feature</th>
							<th>Primary Property</th>
							${bmaReport.detailedAnalysis.comparisonDetails.map(comp => `<th>${comp.address}</th>`).join('')}
						</tr>
					</thead>
					<tbody>
						${bmaReport.detailedAnalysis.featureComparison.map(feature => `
							<tr>
								<td>${feature.feature}</td>
								<td>${feature.primaryValue}</td>
								${feature.comparison.map(comp => `<td>${comp.value}</td>`).join('')}
							</tr>
						`).join('')}
					</tbody>
				</table>

				<h2>Market Trends</h2>
				<p>${bmaReport.detailedAnalysis.marketTrends}</p>

				<div class="recommendation">
					<h2>Recommendation</h2>
					<p>${bmaReport.detailedAnalysis.recommendation}</p>
				</div>
			</body>
			</html>
		`);

		// Wait for content to load before printing
		printWindow.document.close();
		printWindow.onload = function() {
			printWindow.print();
		};
	}
</script>

<div class="bma-report">
	{#if !bmaReport}
		{#if isLoadingBMA}
			<div class="loading">Generating BMA Report...</div>
		{:else if errorMessage}
			<div class="error">{errorMessage}</div>
		{:else}
			<div class="no-properties">
				<p>You need to first add or enable a property for comparison.</p>
				<p>1. Add properties using the Chrome extension</p>
				<p>2. Enable properties for comparison by checking the "Enabled" box</p>
			</div>
		{/if}
	{:else}
		<div class="report-header">
			<h2>BMA Report</h2>
			<div class="header-buttons">
				<button 
					on:click={refreshReport} 
					disabled={isLoadingBMA}
					class="action-button"
				>
					<svg 
						xmlns="http://www.w3.org/2000/svg" 
						width="16" 
						height="16" 
						viewBox="0 0 24 24" 
						fill="none" 
						stroke="currentColor" 
						stroke-width="2" 
						stroke-linecap="round" 
						stroke-linejoin="round"
						class={isLoadingBMA ? 'spinning' : ''}
					>
						<path d="M21 2v6h-6"></path>
						<path d="M3 12a9 9 0 0 1 15-6.7L21 8"></path>
						<path d="M3 22v-6h6"></path>
						<path d="M21 12a9 9 0 0 1-15 6.7L3 16"></path>
					</svg>
					Refresh Report
				</button>
				<button class="action-button" on:click={handleDownload}>
					<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
						<polyline points="7 10 12 15 17 10"></polyline>
						<line x1="12" y1="15" x2="12" y2="3"></line>
					</svg>
					Download PDF
				</button>
			</div>
		</div>

		{#if errorMessage}
			<div class="error-message">
				{errorMessage}
			</div>
		{/if}

		{#if isLoadingBMA}
			<div class="loading">
				Generating report...
			</div>
		{:else if bmaReport}
			<div class="report-content">
				<h3>Summary</h3>
				<p>{bmaReport.summary}</p>
				
				{#if bmaReport.detailedAnalysis}
					<div class="analysis-section">
						<h3>Detailed Analysis</h3>
						
						<div class="property-details">
							<h4>Primary Property Details</h4>
							<div class="details-grid">
								<div class="detail-item">
									<span class="label">Address:</span>
									<span class="value">{bmaReport.detailedAnalysis.primaryPropertyDetails.address}</span>
								</div>
								<div class="detail-item">
									<span class="label">Price:</span>
									<span class="value">${bmaReport.detailedAnalysis.primaryPropertyDetails.price.toLocaleString()}</span>
								</div>
								<div class="detail-item">
									<span class="label">Bedrooms:</span>
									<span class="value">{bmaReport.detailedAnalysis.primaryPropertyDetails.bedrooms}</span>
								</div>
								<div class="detail-item">
									<span class="label">Bathrooms:</span>
									<span class="value">{bmaReport.detailedAnalysis.primaryPropertyDetails.bathrooms}</span>
								</div>
								<div class="detail-item">
									<span class="label">Square Footage:</span>
									<span class="value">{bmaReport.detailedAnalysis.primaryPropertyDetails.squareFootage.toLocaleString()}</span>
								</div>
								<div class="detail-item">
									<span class="label">Year Built:</span>
									<span class="value">{bmaReport.detailedAnalysis.primaryPropertyDetails.yearBuilt}</span>
								</div>
								<div class="detail-item">
									<span class="label">Property Type:</span>
									<span class="value">{bmaReport.detailedAnalysis.primaryPropertyDetails.propertyType}</span>
								</div>
								<div class="detail-item">
									<span class="label">Lot Size:</span>
									<span class="value">{bmaReport.detailedAnalysis.primaryPropertyDetails.lotSize}</span>
								</div>
							</div>
						</div>

						<div class="price-analysis">
							<h4>Price Analysis</h4>
							<p>{bmaReport.detailedAnalysis.priceAnalysis}</p>
						</div>

						<div class="feature-comparison">
							<h4>Feature Comparison</h4>
							{#each bmaReport.detailedAnalysis.featureComparison as feature}
								<div class="feature-item">
									<h5>{feature.feature}</h5>
									<p><strong>Primary Property:</strong> {feature.primaryValue}</p>
									<ul>
										{#each feature.comparison as comp}
											<li><strong>{comp.address}:</strong> {comp.value}</li>
										{/each}
									</ul>
									<p class="analysis-text">{feature.analysis}</p>
								</div>
							{/each}
						</div>

						<div class="market-trends">
							<h4>Market Trends</h4>
							<p>{bmaReport.detailedAnalysis.marketTrends}</p>
						</div>

						<div class="recommendation">
							<h4>Recommendation</h4>
							<p>{bmaReport.detailedAnalysis.recommendation}</p>
						</div>
					</div>
				{/if}
			</div>
		{/if}
	{/if}
</div>

<style>
	.bma-report {
		margin-top: 20px;
		padding: 20px;
		border: 1px solid #ccc;
		border-radius: 4px;
		position: relative;
	}

	.report-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 20px;
	}

	.header-buttons {
		display: flex;
		gap: 8px;
	}

	.action-button {
		display: flex;
		align-items: center;
		gap: 8px;
		background: #2c3e50;
		color: white;
		border: none;
		padding: 8px 16px;
		border-radius: 4px;
		cursor: pointer;
		font-size: 0.9em;
		transition: background-color 0.2s;
	}

	.action-button:hover {
		background: #34495e;
	}

	.action-button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.action-button svg {
		width: 16px;
		height: 16px;
	}

	.action-button svg.spinning {
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}

	.error-message {
		color: #dc3545;
		margin-bottom: 10px;
	}

	.loading {
		text-align: center;
		padding: 20px;
	}

	.report-content {
		line-height: 1.6;
	}

	.report-content h3 {
		margin-top: 20px;
		margin-bottom: 10px;
	}

	.report-content p {
		margin-bottom: 15px;
	}

	.analysis-section {
		margin-top: 20px;
		padding: 15px;
		background-color: white;
		border-radius: 5px;
		box-shadow: 0 1px 3px rgba(0,0,0,0.1);
	}

	.property-details,
	.price-analysis,
	.feature-comparison,
	.market-trends,
	.recommendation {
		margin-bottom: 20px;
		padding: 15px;
		background-color: #f8f9fa;
		border-radius: 4px;
	}

	.details-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 10px;
		margin-top: 10px;
	}

	.detail-item {
		display: flex;
		flex-direction: column;
	}

	.detail-item .label {
		font-weight: bold;
		color: #666;
		margin-bottom: 4px;
	}

	.detail-item .value {
		color: #333;
	}

	.feature-item {
		margin-bottom: 15px;
		padding: 10px;
		background-color: white;
		border-radius: 4px;
		box-shadow: 0 1px 3px rgba(0,0,0,0.1);
	}

	.analysis-text {
		font-style: italic;
		margin-top: 10px;
		color: #666;
	}

	h4 {
		color: #2c3e50;
		margin-bottom: 10px;
	}

	h5 {
		color: #34495e;
		margin-bottom: 5px;
	}

	ul {
		list-style-type: none;
		padding-left: 0;
	}

	li {
		margin-bottom: 5px;
		padding: 5px;
		background-color: #f8f9fa;
		border-radius: 3px;
	}

	.no-properties {
		background: #f8f9fa;
		border: 1px solid #dee2e6;
		border-radius: 8px;
		padding: 20px;
		margin: 20px 0;
		text-align: center;
	}

	.no-properties p {
		margin: 10px 0;
		color: #6c757d;
	}

	.no-properties p:first-child {
		font-weight: 500;
		color: #495057;
	}

	.download-button {
		display: flex;
		align-items: center;
		gap: 8px;
		background: #2c3e50;
		color: white;
		border: none;
		padding: 8px 16px;
		border-radius: 4px;
		cursor: pointer;
		font-size: 0.9em;
		transition: background-color 0.2s;
	}

	.download-button:hover {
		background: #34495e;
	}

	.download-button svg {
		width: 16px;
		height: 16px;
	}
</style> 