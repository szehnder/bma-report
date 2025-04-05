document.addEventListener('DOMContentLoaded', function() {
    const collectButton = document.getElementById('collectButton');
    const statusDiv = document.getElementById('status');

    collectButton.addEventListener('click', async () => {
        try {
            // Get the active tab
            const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
            
            if (!tab.url.startsWith('http')) {
                showStatus('Please navigate to a webpage first', 'error');
                return;
            }

            // Execute the content script in the current tab
            await chrome.scripting.executeScript({
                target: { tabId: tab.id },
                function: collectPageData
            });

            showStatus('Data collection started', 'success');
        } catch (error) {
            showStatus('Error: ' + error.message, 'error');
        }
    });
});

function showStatus(message, type) {
    const statusDiv = document.getElementById('status');
    statusDiv.textContent = message;
    statusDiv.className = type;
    statusDiv.style.display = 'block';
    
    // Hide status after 3 seconds
    setTimeout(() => {
        statusDiv.style.display = 'none';
    }, 3000);
}

// This function will be injected into the page
function collectPageData() {
    const url = window.location.href;
    const content = document.documentElement.innerText || "";

    // Send message to background script
    chrome.runtime.sendMessage({
        type: 'PAGE_DATA',
        data: { url, content }
    }, response => {
        if (chrome.runtime.lastError) {
            console.error('Error:', chrome.runtime.lastError);
            return;
        }
        if (response && response.success) {
            console.log('Data sent successfully');
        } else {
            console.error('Failed to send data');
        }
    });
} 