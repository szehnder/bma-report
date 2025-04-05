<script lang="ts">
	import { onMount } from 'svelte';
	import { API_URL } from '../lib/constants';
	import AddressList from '../components/AddressList.svelte';
	import BMAReport from '../components/BMAReport.svelte';
	import LLMInstructions from '../components/LLMInstructions.svelte';

	let addresses: Array<{
		id: string;
		addressStr: string;
		enabled: boolean;
		primary: boolean;
	}> = [];
	let bmaReport: {
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
	let isLoadingBMA = false;
	let errorMessage = '';
	let llmInstructions = '';

	async function fetchAddresses() {
		try {
			const response = await fetch(`${API_URL}/api/addresses`);
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			addresses = await response.json();
		} catch (err) {
			console.error('Error fetching addresses:', err);
		}
	}

	async function fetchBMAReport() {
		isLoadingBMA = true;
		errorMessage = '';
		
		try {
			const response = await fetch(`${API_URL}/api/bma-report`);
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			bmaReport = await response.json();
		} catch (err) {
			console.error('Error fetching BMA report:', err);
			errorMessage = 'Failed to fetch BMA report. Please try again.';
		} finally {
			isLoadingBMA = false;
		}
	}

	async function fetchLLMInstructions() {
		try {
			const response = await fetch(`${API_URL}/api/llm-instructions`);
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			const data = await response.json();
			llmInstructions = data.instructions || '';
		} catch (err) {
			console.error('Error fetching LLM instructions:', err);
		}
	}

	async function handleInstructionsUpdate() {
		await fetchLLMInstructions();
		await fetchBMAReport();
	}

	onMount(async () => {
		await fetchAddresses();
		await fetchLLMInstructions();
		await fetchBMAReport();
	});
</script>

<main>
	<header>
		<h1>BMA Calculator</h1>
	</header>
	
	<AddressList 
		addresses={addresses} 
		onUpdate={fetchAddresses} 
	/>
	
	<LLMInstructions 
		instructions={llmInstructions} 
		onUpdate={handleInstructionsUpdate} 
	/>
	
	<BMAReport 
		bmaReport={bmaReport} 
		isLoadingBMA={isLoadingBMA} 
		errorMessage={errorMessage} 
	/>
</main>

<style>
	@import url('https://fonts.googleapis.com/css2?family=Roboto:wght@300;400;500;700&display=swap');

	main {
		max-width: 800px;
		margin: 0 auto;
		padding: 20px;
		font-family: 'Roboto', sans-serif;
	}

	header {
		background: #2c3e50;
		padding: 20px;
		margin: -20px -20px 20px -20px;
		border-radius: 4px 4px 0 0;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
	}

	h1 {
		margin: 0;
		color: white;
		font-weight: 500;
		font-size: 1.8em;
		letter-spacing: 0.5px;
	}
</style>
