document.addEventListener('DOMContentLoaded', () => {
    const codeBlock = document.querySelector('.paste-content pre code');
    const lineNumbersDiv = document.querySelector('.line-numbers-container .line-numbers');
    const lineCountSpan = document.getElementById('lineCount');
    const searchInput = document.getElementById('searchInput');
    const searchButton = document.getElementById('searchButton');
    const prevMatchButton = document.getElementById('prevMatchButton');
    const nextMatchButton = document.getElementById('nextMatchButton');
    const searchResultCount = document.getElementById('searchResultCount');
    const toggleWrapButton = document.getElementById('toggleWrap');
    const copyButton = document.getElementById('copyButton');
    const copyShareLinkButton = document.getElementById('copyShareLinkButton');
    const shareLinkInput = document.getElementById('shareLinkInput');
    const timeRemainingSpan = document.getElementById('time-remaining');
    const pasteContentContainer = document.querySelector('.paste-content');

    // Exit early if essential elements are missing (e.g., on index or error page)
    if (!codeBlock) {
        // If it's the index page, only the shareLinkInput might exist for paste.html, so
        // don't exit if we're on the index page and don't expect a codeBlock.
        // For paste.html, codeBlock is essential.
        if (window.location.pathname.length > 1 && window.location.pathname !== '/password_prompt.html') {
            return;
        }
    }

    // --- Line Numbers ---
    // The line count logic is good. It correctly splits by newline and updates the div.
    // The scroll synchronization is also correctly implemented by attaching listeners to codeBlock.
    const updateLineNumbers = () => {
        const content = codeBlock.textContent;
        const lines = content.split('\n');
        lineNumbersDiv.innerHTML = '';
        let numbersHtml = '';
        for (let i = 1; i <= lines.length; i++) {
            numbersHtml += `<span>${i}</span>\n`;
        }
        lineNumbersDiv.innerHTML = numbersHtml;
        lineCountSpan.textContent = `Lines: ${lines.length}`;

        // Sync scrollbar for line numbers and content
        // Attach event listeners only once to avoid multiple listeners on subsequent calls if `updateLineNumbers` was called multiple times.
        // It's already attached only once on DOMContentLoaded for this implementation.
        codeBlock.addEventListener('scroll', () => {
            lineNumbersDiv.scrollTop = codeBlock.scrollTop;
        });
        lineNumbersDiv.addEventListener('scroll', () => {
            codeBlock.scrollTop = lineNumbersDiv.scrollTop;
        });
    };
    updateLineNumbers();

    // --- Syntax Highlighting (Highlight.js) ---
    // The logic is correct. hljs.highlightElement(codeBlock) is the appropriate call for a single block.
    if (typeof hljs !== 'undefined') {
        hljs.highlightElement(codeBlock);
    }

    // --- Search Functionality ---
    let originalContent = codeBlock.textContent;
    let matches = [];
    let currentMatchIndex = -1;

    const performSearch = () => {
        const searchTerm = searchInput.value;
        if (!searchTerm) {
            codeBlock.innerHTML = originalContent;
            searchResultCount.textContent = '';
            matches = [];
            currentMatchIndex = -1;
            prevMatchButton.disabled = true;
            nextMatchButton.disabled = true;
            if (typeof hljs !== 'undefined') hljs.highlightElement(codeBlock);
            return;
        }

        // Rebuild regex, replace existing highlights, re-highlight the entire block
        const regex = new RegExp(searchTerm.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi');
        let highlightedContent = originalContent.replace(regex, (match) => `<span class="highlight">${match}</span>`);

        codeBlock.innerHTML = highlightedContent;
        if (typeof hljs !== 'undefined') hljs.highlightElement(codeBlock);

        // Re-find all matches in the original content for navigation
        matches = [...originalContent.matchAll(regex)];
        searchResultCount.textContent = `${matches.length} results`;

        if (matches.length > 0) {
            currentMatchIndex = 0;
            scrollToMatch(currentMatchIndex);
            prevMatchButton.disabled = false;
            nextMatchButton.disabled = false;
        } else {
            currentMatchIndex = -1;
            searchResultCount.textContent = '0 results';
            prevMatchButton.disabled = true;
            nextMatchButton.disabled = true;
        }
    };

    const scrollToMatch = (index) => {
        if (matches.length === 0 || index === -1) return;

        const allHighlights = codeBlock.querySelectorAll('.highlight');
        if (allHighlights.length > 0) {
            // Remove active class from previously active highlight
            const currentActive = codeBlock.querySelector('.highlight.active');
            if (currentActive) {
                currentActive.classList.remove('active');
            }

            // Add active class to the target highlight
            const targetHighlight = allHighlights[index];
            targetHighlight.classList.add('active');

            // Scroll the code block to make the target highlight visible
            targetHighlight.scrollIntoView({ behavior: 'smooth', block: 'center' });

            searchResultCount.textContent = `${index + 1}/${matches.length} results`;
        }
    };

    // Search Navigation Logic
    const navigateSearch = (direction) => {
        if (matches.length === 0) return;

        let newIndex = currentMatchIndex + direction;

        if (newIndex < 0) {
            newIndex = matches.length - 1; // Wrap around to end
        } else if (newIndex >= matches.length) {
            newIndex = 0; // Wrap around to beginning
        }

        currentMatchIndex = newIndex;
        scrollToMatch(currentMatchIndex);
    };

    // Event Listeners for Search and Navigation
    searchButton.addEventListener('click', performSearch);
    prevMatchButton.addEventListener('click', () => navigateSearch(-1));
    nextMatchButton.addEventListener('click', () => navigateSearch(1));

    searchInput.addEventListener('keyup', (e) => {
        if (e.key === 'Enter') {
            performSearch();
        }
        // No live search preview here, as noted it can be performance intensive
    });

    // Initialize nav buttons as disabled on page load
    prevMatchButton.disabled = true;
    nextMatchButton.disabled = true;

    // --- Toggle Word Wrap ---
    toggleWrapButton.addEventListener('click', () => {
        pasteContentContainer.classList.toggle('wrap-enabled');
    });

    // --- Copy Raw Content ---
    // Helper for button feedback
    const setButtonFeedback = (buttonElement, originalText, successText, errorText, isSuccess) => {
        const textSpan = buttonElement.querySelector('.button-text-feedback');
        if (!textSpan) return;

        const originalIcon = buttonElement.querySelector('svg');
        buttonElement.classList.add('feedback-active');
        if (originalIcon) originalIcon.style.display = 'none';

        textSpan.textContent = isSuccess ? successText : errorText;
        textSpan.style.color = isSuccess ? 'var(--green)' : 'var(--red)';

        setTimeout(() => {
            textSpan.textContent = originalText;
            textSpan.style.color = '';
            if (originalIcon) originalIcon.style.display = '';
            buttonElement.classList.remove('feedback-active');
        }, 2000);
    };

    if (copyButton) {
        copyButton.addEventListener('click', () => {
            navigator.clipboard.writeText(originalContent)
                .then(() => {
                    setButtonFeedback(copyButton, 'Copy Raw', 'Copied!', 'Error!', true);
                })
                .catch(err => {
                    console.error('Failed to copy text: ', err);
                    setButtonFeedback(copyButton, 'Copy Raw', 'Copied!', 'Error!', false);
                });
        });
    }

    // --- Expiration Countdown ---
    if (timeRemainingSpan) {
        const expiryTimeStr = timeRemainingSpan.dataset.expiry;
        const expiryTime = new Date(expiryTimeStr);

        const updateCountdown = () => {
            const now = new Date();
            const diff = expiryTime.getTime() - now.getTime();

            if (diff <= 0) {
                timeRemainingSpan.textContent = 'Expired';
                clearInterval(countdownInterval);
                return;
            }

            const days = Math.floor(diff / (1000 * 60 * 60 * 24));
            const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
            const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
            const seconds = Math.floor((diff % (1000 * 60)) / 1000);

            let countdownText = [];
            if (days > 0) countdownText.push(`${days}d`);
            if (hours > 0) countdownText.push(`${hours}h`);
            if (minutes > 0) countdownText.push(`${minutes}m`);
            if (seconds > 0 || countdownText.length === 0) countdownText.push(`${seconds}s`);

            timeRemainingSpan.textContent = countdownText.join(' ');
        };

        updateCountdown();
        const countdownInterval = setInterval(updateCountdown, 1000);
    }

    // --- Copy Share Link ---
    // The previous implementation used innerHTML manipulation for SVG icons.
    // Let's adapt it to use the new setButtonFeedback helper.
    if (copyShareLinkButton && shareLinkInput) {
        copyShareLinkButton.addEventListener('click', () => {
            shareLinkInput.select();
            shareLinkInput.setSelectionRange(0, 99999);

            navigator.clipboard.writeText(shareLinkInput.value)
                .then(() => {
                    setButtonFeedback(copyShareLinkButton, 'Copy Link', 'Copied!', 'Error!', true);
                })
                .catch(err => {
                    console.error('Failed to copy share link: ', err);
                    setButtonFeedback(copyShareLinkButton, 'Copy Link', 'Copied!', 'Error!', false);
                });
        });
    }

    // Initial highlight (ensure this is called after line numbers and all setup)
    // This is already being called once after DOMContentLoaded listener.
    // hljs.highlightElement(codeBlock) is called within performSearch too.
    // No redundant call needed here if performSearch or page load causes initial highlight.
    // However, if no search is performed, this ensures initial highlight.
    if (typeof hljs !== 'undefined' && codeBlock) { // Check codeBlock again to be safe
        hljs.highlightElement(codeBlock);
    }
});