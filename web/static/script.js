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

    if (!codeBlock && window.location.pathname.length > 1 && window.location.pathname !== '/password_prompt.html') {
        return;
    }

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

        codeBlock.addEventListener('scroll', () => {
            lineNumbersDiv.scrollTop = codeBlock.scrollTop;
        });
        lineNumbersDiv.addEventListener('scroll', () => {
            codeBlock.scrollTop = lineNumbersDiv.scrollTop;
        });
    };
    if (codeBlock) updateLineNumbers();

    if (typeof hljs !== 'undefined' && codeBlock) {
        hljs.highlightElement(codeBlock);
    }

    let originalContent = codeBlock ? codeBlock.textContent : '';
    let matches = [];
    let currentMatchIndex = -1;

    const performSearch = () => {
        const searchTerm = searchInput.value;
        if (!searchTerm) {
            codeBlock.innerHTML = originalContent;
            searchResultCount.textContent = '';
            matches = [];
            currentMatchIndex = -1;
            if (prevMatchButton) prevMatchButton.disabled = true;
            if (nextMatchButton) nextMatchButton.disabled = true;
            if (typeof hljs !== 'undefined') hljs.highlightElement(codeBlock);
            return;
        }

        const regex = new RegExp(searchTerm.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi');
        let highlightedContent = originalContent.replace(regex, (match) => `<span class="highlight">${match}</span>`);

        codeBlock.innerHTML = highlightedContent;
        if (typeof hljs !== 'undefined') hljs.highlightElement(codeBlock);

        matches = [...originalContent.matchAll(regex)];
        searchResultCount.textContent = `${matches.length} results`;

        if (matches.length > 0) {
            currentMatchIndex = 0;
            scrollToMatch(currentMatchIndex);
            if (prevMatchButton) prevMatchButton.disabled = false;
            if (nextMatchButton) nextMatchButton.disabled = false;
        } else {
            currentMatchIndex = -1;
            searchResultCount.textContent = '0 results';
            if (prevMatchButton) prevMatchButton.disabled = true;
            if (nextMatchButton) nextMatchButton.disabled = true;
        }
    };

    const scrollToMatch = (index) => {
        if (matches.length === 0 || index === -1) return;

        const allHighlights = codeBlock.querySelectorAll('.highlight');
        if (allHighlights.length > 0) {
            const currentActive = codeBlock.querySelector('.highlight.active');
            if (currentActive) {
                currentActive.classList.remove('active');
            }

            const targetHighlight = allHighlights[index];
            targetHighlight.classList.add('active');
            targetHighlight.scrollIntoView({ behavior: 'smooth', block: 'center' });

            searchResultCount.textContent = `${index + 1}/${matches.length} results`;
        }
    };

    const navigateSearch = (direction) => {
        if (matches.length === 0) return;
        let newIndex = currentMatchIndex + direction;
        if (newIndex < 0) newIndex = matches.length - 1;
        else if (newIndex >= matches.length) newIndex = 0;
        currentMatchIndex = newIndex;
        scrollToMatch(currentMatchIndex);
    };

    if (searchButton) searchButton.addEventListener('click', performSearch);
    if (prevMatchButton) prevMatchButton.addEventListener('click', () => navigateSearch(-1));
    if (nextMatchButton) nextMatchButton.addEventListener('click', () => navigateSearch(1));
    if (searchInput) {
        searchInput.addEventListener('keyup', (e) => {
            if (e.key === 'Enter') performSearch();
        });
    }

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
                .catch(() => {
                    setButtonFeedback(copyButton, 'Copy Raw', 'Copied!', 'Error!', false);
                });
        });
    }

    if (toggleWrapButton) {
        toggleWrapButton.addEventListener('click', () => {
            pasteContentContainer.classList.toggle('wrap-enabled');
        });
    }

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

    if (copyShareLinkButton && shareLinkInput) {
        copyShareLinkButton.addEventListener('click', () => {
            shareLinkInput.select();
            shareLinkInput.setSelectionRange(0, 99999);

            navigator.clipboard.writeText(shareLinkInput.value)
                .then(() => {
                    setButtonFeedback(copyShareLinkButton, 'Copy Link', 'Copied!', 'Error!', true);
                })
                .catch(() => {
                    setButtonFeedback(copyShareLinkButton, 'Copy Link', 'Copied!', 'Error!', false);
                });
        });
    }

    if (typeof hljs !== 'undefined' && codeBlock) {
        hljs.highlightElement(codeBlock);
    }
});
