import 'webextension-polyfill';

// Handle extension icon click
chrome.action.onClicked.addListener(async tab => {
  // Open the side panel
  await chrome.sidePanel.open({ windowId: tab.windowId });
});

// Make side panel persist across all sites
chrome.sidePanel.setPanelBehavior({ openPanelOnActionClick: true });

// Set the side panel to be available on all URLs
chrome.tabs.onUpdated.addListener((tabId, info) => {
  if (info.status === 'complete') {
    chrome.sidePanel.setOptions({
      tabId,
      path: 'side-panel/index.html',
      enabled: true,
    });
  }
});
