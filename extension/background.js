// Listen for extension installation
chrome.runtime.onInstalled.addListener(() => {
    console.log("BMA Address Collector installed");
});

// Listen for messages from content script
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.type === 'PAGE_DATA') {
        const { url, content } = message.data;
        
        // Send data to backend
        fetch("http://localhost:8080/api/extension/page-data", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ url, content })
        })
        .then(response => response.json())
        .then(data => {
            console.log("Data sent to backend successfully:", data);
            sendResponse({ success: true, data });
        })
        .catch(error => {
            console.error("Failed to send data to backend:", error);
            sendResponse({ success: false, error: error.message });
        });

        // Return true to indicate we'll send a response asynchronously
        return true;
    }
}); 