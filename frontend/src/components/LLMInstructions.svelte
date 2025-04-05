<script lang="ts">
	import { API_URL } from '../lib/constants';

	export let instructions = '';
	export let onUpdate: () => Promise<void>;

	let tempInstructions = '';
	let errorMessage = '';
	let isSaving = false;

	async function saveInstructions() {
		if (isSaving) return;
		isSaving = true;
		errorMessage = '';

		try {
			const response = await fetch(`${API_URL}/api/llm-instructions`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ instructions: tempInstructions })
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			instructions = tempInstructions;
			await onUpdate();
		} catch (err) {
			console.error('Error saving instructions:', err);
			errorMessage = 'Failed to save instructions. Please try again.';
		} finally {
			isSaving = false;
		}
	}

	$: tempInstructions = instructions;
</script>

<div class="llm-instructions">
	<h3>LLM Instructions</h3>
	
	{#if errorMessage}
		<div class="error-message">
			{errorMessage}
		</div>
	{/if}

	<textarea
		bind:value={tempInstructions}
		placeholder="Enter additional instructions for the LLM..."
		rows="4"
	></textarea>
	
	<button 
		on:click={saveInstructions} 
		class="save-button"
		disabled={isSaving}
	>
		{#if isSaving}
			Saving...
		{:else}
			Save LLM Instructions
		{/if}
	</button>
</div>

<style>
	.llm-instructions {
		margin: 20px 0;
		padding: 15px;
		background-color: #f8f9fa;
		border-radius: 4px;
	}

	textarea {
		width: 100%;
		padding: 8px;
		margin-bottom: 10px;
		border: 1px solid #ccc;
		border-radius: 4px;
		resize: vertical;
	}

	.save-button {
		padding: 8px 16px;
		background-color: #28a745;
		color: white;
		border: none;
		border-radius: 4px;
		cursor: pointer;
		transition: background-color 0.2s;
	}

	.save-button:hover:not(:disabled) {
		background-color: #218838;
	}

	.save-button:disabled {
		background-color: #6c757d;
		cursor: not-allowed;
	}

	.error-message {
		color: #dc3545;
		margin-bottom: 10px;
	}
</style> 