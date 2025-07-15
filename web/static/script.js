document.addEventListener('DOMContentLoaded', () => {
	if (document.querySelector('.form-component')) {
		initHomePage();
	} else if (document.querySelector('.paste-component')) {
		initPastePage();
	}

	// --- Sakura Animation Toggle ---
	const sakuraToggle = document.getElementById('cursor-toggle');
	const sakuraContainer = document.getElementById('sakura-container');
	let sakuraEnabled = false;
	function createSakuraPetals(count = 15) {
		sakuraContainer.innerHTML = '';
		for (let i = 0; i < count; i++) {
			const petal = document.createElement('img');
			petal.src = '/static/sakura.png';
			petal.className = 'sakura-petal';
	        petal.style.left = Math.random() * 100 + 'vw';
	        petal.style.top = Math.random() * -100 + 'px';
	        petal.style.animationDelay = (Math.random() * 10) + 's';
			sakuraContainer.appendChild(petal);
		}
	}
	if (sakuraContainer) sakuraContainer.style.display = 'none';
	if (sakuraToggle) sakuraToggle.classList.remove('active');
	if (sakuraToggle) {
		sakuraToggle.onclick = () => {
			sakuraEnabled = !sakuraEnabled;
			if (sakuraContainer) {
				sakuraContainer.style.display = sakuraEnabled ? 'block' : 'none';
			}
			sakuraToggle.classList.toggle('active', sakuraEnabled);
			if (sakuraEnabled) {
				createSakuraPetals();
			} else {
				sakuraContainer.innerHTML = '';
			}
		};
	}
});

function initHomePage() {
	const textInput = document.getElementById('content-input');
	const languageSelect = document.getElementById('language');
	const charCounter = document.getElementById('char-counter');
	const detectedLanguage = document.getElementById('detected-language');
	const tabs = document.querySelectorAll('.tab-button');
	const tabPanels = document.querySelectorAll('.tab-panel');
	const form = document.querySelector('.form-component form');
	const fileInput = document.getElementById('file-input');
	const dropZone = document.querySelector('.drop-zone');

	// Initialize state asynchronously
	initializePageState();

	async function initializePageState() {
		const content = document.getElementById("content");
		const draftKey = "paste_draft_content";
		const languageKey = "selectedLanguage";

		// Initialize content and language state
		if (content) {
			const savedContent = localStorage.getItem(draftKey) || "";
			const savedLanguage = localStorage.getItem(languageKey);

			// Restore saved content
			content.value = savedContent;

			// Update character counter immediately for restored content
			updateCharCounter();

			// Smart language selection logic
			if (languageSelect) {
				const navEntries = performance.getEntriesByType('navigation');
				const isReload = navEntries.length && navEntries[0].type === 'reload';

				if (isReload) {
					// Explicit reload - reset everything to clean state
					languageSelect.value = 'auto';
					localStorage.removeItem(languageKey);
					updateDetectedLanguage('');
				} else if (savedContent.trim() && savedLanguage) {
					// Has saved content AND saved language - restore both
					if (languageSelect.querySelector(`option[value="${savedLanguage}"]`)) {
						languageSelect.value = savedLanguage;
						if (savedLanguage !== 'auto') {
							updateDetectedLanguage('');
						}
					}
				} else if (savedContent.trim() && !savedLanguage) {
					// Has saved content but NO saved language - auto-detect
					languageSelect.value = 'auto';
					const detected = await detectLanguage(savedContent);
					if (detected) {
						updateDetectedLanguage(detected);
					}
				} else {
					// No saved content - default to auto
					languageSelect.value = 'auto';
					updateDetectedLanguage('');
				}
			}

			// Save content changes
			content.addEventListener("input", () => {
				localStorage.setItem(draftKey, content.value);
			});

			// Focus the content area
			content.focus();
		}
	}

	// Character counter
	function updateCharCounter() {
		const content = document.getElementById('content-input') || document.getElementById('content');
		if (content && charCounter) {
			const count = content.value.length;
			charCounter.textContent = `${count.toLocaleString()} characters`;

			// Show warning for very large pastes
			if (count > 100000) {
				charCounter.style.color = 'var(--yellow)';
			} else if (count > 500000) {
				charCounter.style.color = 'var(--red)';
			} else {
				charCounter.style.color = 'var(--text-secondary)';
			}
		}
	}

	// Update detected language display
	function updateDetectedLanguage(lang) {
		if (detectedLanguage) {
			if (lang && lang !== 'auto' && lang !== 'plaintext') {
				detectedLanguage.textContent = `Detected: ${lang}`;
			} else {
				detectedLanguage.textContent = '';
			}
		}
	}

	async function loadLang(lang) {
		// Skip if already loaded
		if (hljs.getLanguage(lang)) {
			return;
		}

		try {
			// Dynamically import the language file
			const script = document.createElement('script');
			script.src = `/static/hj/languages/${lang}.min.js`;
			script.async = true;

			// Return a promise that resolves when the script loads
			return new Promise((resolve, reject) => {
				script.onload = () => resolve();
				script.onerror = () => {
					console.warn(`Failed to load language: ${lang}`);
					reject(new Error(`Failed to load ${lang}`));
				};
				document.head.appendChild(script);
			});
		} catch (error) {
			console.warn(`Language ${lang} not available:`, error);
		}
	}

	if (languageSelect) {
		const navEntries = performance.getEntriesByType('navigation');
		const isReload = navEntries.length && navEntries[0].type === 'reload';

		if (isReload) {
			languageSelect.value = 'auto';
			localStorage.removeItem('selectedLanguage');
		} else {
			const savedLang = localStorage.getItem('selectedLanguage');
			if (savedLang && languageSelect.querySelector(`option[value="${savedLang}"]`)) {
				languageSelect.value = savedLang;
			}
		}

		languageSelect.addEventListener('change', () => {
			const val = languageSelect.value;
			if (val === 'auto') {
				localStorage.removeItem('selectedLanguage');
			} else {
				localStorage.setItem('selectedLanguage', val);
			}
		});
	}

	// Lightweight language detection without loading all languages
	async function detectLanguage(text) {
		// Common language patterns for quick detection without full language files
		const patterns = {
			'javascript': [/\b(function|const|let|var|class|export|import)\b/, /=>/, /\bthis\./],
			'typescript': [/\b(interface|type|extends|implements)\b/, /:.*\[|\]/, /\bas\s+\w+/],
			'python': [/\b(def|class|import|from|if __name__)\b/, /:\s*$/, /\bself\./],
			'java': [/\b(public|private|class|static|void)\b/, /\w+\s*\(.*\)\s*{/, /System\.out\.print/],
			'cpp': [/\b(#include|using namespace|std::)\b/, /\w+::\w+/, /cout\s*<</],
			'c': [/\b(#include|int main|printf|malloc)\b/, /\w+\s*\*\w+/, /scanf/],
			'go': [/\b(package|import|func|var|type)\b/, /:=/, /fmt\.Print/],
			'rust': [/\b(fn|let|mut|use|struct|impl)\b/, /\w+!/, /println!/],
			'php': [/<\?php/, /\$\w+/, /echo\s+/],
			'ruby': [/\b(def|class|module|end|puts)\b/, /@\w+/, /\w+\.each/],
			'bash': [/\b(if.*then|for.*do|while.*do)\b/, /^\s*#!/, /\$\w+/, /echo\s+/],
			'sql': [/\b(SELECT|FROM|WHERE|INSERT|UPDATE|DELETE)\b/i, /\bJOIN\b/i],
			'css': [/\w+\s*{/, /:\s*[\w#%]+;/, /@media/],
			'html': [/<\w+/, /<\/\w+>/, /<!DOCTYPE/],
			'json': [/^\s*{/, /"\w+":\s*/, /^\s*\[/],
			'xml': [/<\?xml/, /<\w+.*>/, /<\/\w+>/],
			'yaml': [/^\w+:/, /^\s*-\s+/, /---/],
			'dockerfile': [/\b(FROM|RUN|COPY|CMD)\b/, /\bEXPOSE\s+\d+/],
			'makefile': [/^\w+:/, /\t@?/, /\$\(.*\)/]
		};

		// Score each language based on pattern matches
		const scores = {};
		for (const [lang, langPatterns] of Object.entries(patterns)) {
			scores[lang] = 0;
			for (const pattern of langPatterns) {
				if (pattern.test(text)) {
					scores[lang]++;
				}
			}
		}

		// Find the language with the highest score
		const detected = Object.entries(scores)
			.filter(([_, score]) => score > 0)
			.sort(([, a], [, b]) => b - a)[0];

		return detected ? detected[0] : null;
	}

	// auto-detect language on paste and real-time detection
	if (textInput && languageSelect && typeof hljs !== 'undefined') {
		// Initialize character counter
		updateCharCounter();
		textInput.addEventListener('input', updateCharCounter);

		// Language detection on paste
		textInput.addEventListener('paste', async (e) => {
			if (languageSelect.value !== 'auto') return;
			const pastedText = (e.clipboardData || window.clipboardData).getData('text');

			// Use lightweight detection first
			const detected = await detectLanguage(pastedText);
			if (detected && languageSelect.querySelector(`option[value="${detected}"]`)) {
				languageSelect.value = detected;
				localStorage.setItem('selectedLanguage', detected);
				updateDetectedLanguage(detected);
			}

			// Update character counter after paste
			setTimeout(updateCharCounter, 10);
		});

		// Real-time language detection for auto mode
		let detectTimeout;
		textInput.addEventListener('input', () => {
			clearTimeout(detectTimeout);
			detectTimeout = setTimeout(async () => {
				if (languageSelect.value === 'auto' && textInput.value.trim()) {
					const detected = await detectLanguage(textInput.value);
					if (detected) {
						updateDetectedLanguage(detected);
					} else {
						updateDetectedLanguage('');
					}
				}
			}, 500);
		});

		// Clear detected language when manually selecting
		languageSelect.addEventListener('change', () => {
			if (languageSelect.value !== 'auto') {
				updateDetectedLanguage('');
			}
		});
	}

	// Form submission with better UX
	if (form) {
		const submitButton = form.querySelector('button[type="submit"]');

		form.addEventListener('submit', (e) => {
			if (submitButton) {
				submitButton.classList.add('loading');
				submitButton.disabled = true;
				submitButton.textContent = 'Creating...';
			}

			// Basic validation feedback
			const textContent = textInput ? textInput.value.trim() : '';
			const fileSelected = fileInput && fileInput.files && fileInput.files.length > 0;

			if (!textContent && !fileSelected) {
				e.preventDefault();
				if (submitButton) {
					submitButton.classList.remove('loading');
					submitButton.disabled = false;
					submitButton.textContent = 'Create Paste';
				}

				// Show error feedback
				const activePanel = document.querySelector('.tab-panel.active');
				if (activePanel && activePanel.id === 'text-panel' && textInput) {
					textInput.focus();
					textInput.closest('.tab-panel').style.borderColor = 'var(--red)';
					setTimeout(() => {
						textInput.closest('.tab-panel').style.borderColor = '';
					}, 3000);
				}
			}
		});
	}

	// tab logic
	tabs.forEach(tab => {
		tab.addEventListener('click', () => {
			tabs.forEach(t => t.classList.remove('active'));
			tab.classList.add('active');
			tabPanels.forEach(panel => panel.classList.toggle('active', panel.id === `${tab.dataset.tab}-panel`));
		});
	});

	// drag-and-drop file preview
	if (dropZone && fileInput) {
		const dropZonePrompt = document.querySelector('.drop-zone-prompt');
		const dropZonePreview = document.querySelector('.drop-zone-preview');

		const updateDropZone = (file) => {
			dropZonePrompt.style.display = 'none';
			dropZonePreview.style.display = 'block';
			dropZonePreview.innerHTML = `<strong>${file.name}</strong><small>${(file.size / 1024 / 1024).toFixed(2)} MB</small>`;
		};

		fileInput.addEventListener('change', () => {
			if (fileInput.files.length > 0) updateDropZone(fileInput.files[0]);
		});

		dropZone.addEventListener('dragover', e => {
			e.preventDefault();
			dropZone.classList.add('drag-over');
		});
		['dragleave', 'dragend'].forEach(type => {
			dropZone.addEventListener(type, () => dropZone.classList.remove('drag-over'));
		});
		dropZone.addEventListener('drop', e => {
			e.preventDefault();
			dropZone.classList.remove('drag-over');
			if (e.dataTransfer.files.length > 0) {
				fileInput.files = e.dataTransfer.files;
				updateDropZone(fileInput.files[0]);
			}
		});
	}

	form.addEventListener('submit', async () => {
		const activeTab = document.querySelector('.tab-button.active').dataset.tab;

		if (activeTab === 'text') {
			fileInput.value = '';
			const filePasswordField = document.getElementById('file-password');
			if (filePasswordField) filePasswordField.value = '';

			if (languageSelect.value === 'auto' && textInput.value.trim() !== '') {
				const detected = await detectLanguage(textInput.value);
				if (detected && languageSelect.querySelector(`option[value="${detected}"]`)) {
					languageSelect.value = detected;
					localStorage.setItem('selectedLanguage', detected);
				} else {
					languageSelect.value = 'plaintext';
				}
			}
		} else {
			textInput.value = '';
			const textPasswordField = document.getElementById('password');
			if (textPasswordField) textPasswordField.value = '';
		}
	});

	const content = document.getElementById("content");
	const password = document.getElementById("password");
	if (content) content.focus();

	content?.addEventListener("keydown", (e) => {
		if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
			form?.submit();
		}
	});

	if (content) {
		content.addEventListener("keydown", (e) => {
			if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
				form?.submit();
			}
		});
	}

	form?.addEventListener("submit", () => {
		localStorage.removeItem("paste_draft_content");
	});

	const toggle = document.getElementById("toggle-password");
	const passwordField = document.getElementById("password");
	if (passwordField && toggle) {
		toggle.addEventListener("click", () => {
			const type = passwordField.type === "password" ? "text" : "password";
			passwordField.type = type;
			toggle.textContent = type === "text" ? "hide" : "show";
		});
	}

	const fileToggle = document.getElementById("toggle-file-password");
	const filePasswordField = document.getElementById("file-password");
	if (filePasswordField && fileToggle) {
		fileToggle.addEventListener("click", () => {
			const type = filePasswordField.type === "password" ? "text" : "password";
			filePasswordField.type = type;
			fileToggle.textContent = type === "text" ? "hide" : "show";
		});
	}

	// Sakura mascot popup logic
	const mascotPopup = document.getElementById('sakura-mascot-popup');
	const mascotMessage = document.getElementById('mascot-message');
	const sakuraToggle = document.getElementById('cursor-toggle');

	const mascotMessages = [
	    "Hey there! 🌸 Paste something cool!",
	    "Did you know? Yoru loves privacy!",
	    "Try the sakura animation! Click the flower!",
	    "You can paste files too, not just text!",
	    "Catppuccin theme is the best, right?",
	    "Stay hydrated and write code! 💧",
	    "Share your paste with friends!",
	    "Yoru is open source! Check GitHub!",
	    "Paste expires? Set it to never!",
	    "You can search inside your paste! 🔍"
	];
	let mascotIndex = 0;
	let mascotTypingTimeout = null;
	let mascotCycleTimeout = null;

	function typeMascotMessage(msg, i = 0) {
	    mascotMessage.textContent = '';
	    if (i < msg.length) {
	        mascotMessage.textContent = msg.slice(0, i + 1);
	        mascotTypingTimeout = setTimeout(() => typeMascotMessage(msg, i + 1), 40 + Math.random() * 40);
	    } else {
	        mascotTypingTimeout = null;
	    }
	}

	function showMascotPopup() {
	    if (mascotPopup) mascotPopup.style.display = 'flex';
	    typeMascotMessage(mascotMessages[mascotIndex]);
	    mascotCycleTimeout = setTimeout(nextMascotMessage, 5000);
	}

	function nextMascotMessage() {
	    mascotIndex = (mascotIndex + 1) % mascotMessages.length;
	    typeMascotMessage(mascotMessages[mascotIndex]);
	    mascotCycleTimeout = setTimeout(nextMascotMessage, 5000 + Math.random() * 2000);
	}

	function hideMascotPopup() {
	    if (mascotPopup) mascotPopup.style.display = 'none';
	    if (mascotTypingTimeout) clearTimeout(mascotTypingTimeout);
	    if (mascotCycleTimeout) clearTimeout(mascotCycleTimeout);
	}

	// Only show on index page
	if (mascotPopup && mascotMessage && document.querySelector('.form-component')) {
	    setTimeout(showMascotPopup, 2500); // show after 2.5s
	    // Hide on paste creation or navigation
	    document.querySelector('form')?.addEventListener('submit', hideMascotPopup);
	    // Clicking the mascot bubble cycles messages
	    mascotPopup.addEventListener('click', () => {
	        mascotIndex = Math.floor(Math.random() * mascotMessages.length);
	        typeMascotMessage(mascotMessages[mascotIndex]);
	    });
	}
}



async function initPastePage() {
	const copyToClipboard = (button, textToCopy, originalText) => {
		navigator.clipboard.writeText(textToCopy)
			.then(() => { button.textContent = 'Copied!'; })
			.catch(err => { button.textContent = 'Error!'; console.error('Clipboard error:', err); })
			.finally(() => { setTimeout(() => { button.textContent = originalText; }, 2000); });
	};
	const searchInput = document.getElementById('searchInput');
	const searchButton = document.getElementById('searchButton');

	// make enter key not reload page
	if (searchInput) {
		searchInput.addEventListener('keydown', (e) => {
			if (e.key === 'Enter') {
				e.preventDefault();
				e.stopPropagation();
				searchButton?.click();
			}
		});
	}

	const copyShareLinkButtonSecondary = document.getElementById('copyShareLinkButtonSecondary');
	const shareLinkInput = document.getElementById('shareLinkInput');
	const timeRemainingSpan = document.getElementById('time-remaining');

	// Handle copy button next to input
	if (copyShareLinkButtonSecondary && shareLinkInput) {
		copyShareLinkButtonSecondary.addEventListener('click', () => {
			copyToClipboard(copyShareLinkButtonSecondary, shareLinkInput.value, 'Copy');
		});

		// Auto-select link when clicked for easy copying
		shareLinkInput.addEventListener('click', () => {
			shareLinkInput.select();
		});
	}

	if (timeRemainingSpan && timeRemainingSpan.dataset.expiry) {
		const expiryTime = new Date(timeRemainingSpan.dataset.expiry);
		const countdownInterval = setInterval(() => {
			const diff = expiryTime.getTime() - new Date().getTime();
			if (diff <= 0) {
				timeRemainingSpan.textContent = 'Expired';
				clearInterval(countdownInterval);
				return;
			}
			const d = Math.floor(diff / (1000 * 60 * 60 * 24));
			const h = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
			const m = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
			const s = Math.floor((diff % (1000 * 60)) / 1000);
			timeRemainingSpan.textContent = [d > 0 ? `${d}d` : '', h > 0 ? `${h}h` : '', m > 0 ? `${m}m` : '', (d > 0 || h > 0 || m > 0) ? '' : `${s}s`].filter(Boolean).join(' ') || `${s}s`;
		}, 1000);
	}

	const codeBlock = document.querySelector('.code-view-container pre code');
	if (codeBlock) {
		// Get original content from the code block
		const originalContent = codeBlock.textContent || '';

		const lineNumbersDiv = document.querySelector('.line-numbers');
		const lineCountSpan = document.getElementById('lineCount');
		const codeViewContainer = document.querySelector('.code-view-container');

		const lines = originalContent.split('\n');
		lineCountSpan.textContent = lines.length;
		lineNumbersDiv.innerHTML = Array.from({ length: lines.length }, (_, i) => `<span>${i + 1}</span>`).join('');

		const copyRawButton = document.getElementById('copyButton');
		const toggleWrapButton = document.getElementById('toggleWrap');
		const lineInfoDiv = document.querySelector('.line-info');
		
		copyRawButton.addEventListener('click', () => copyToClipboard(copyRawButton, originalContent, 'Copy Raw'));
		toggleWrapButton.addEventListener('click', () => {
			codeViewContainer.classList.toggle('wrap-enabled');
			// Only hide the line count info in footer, keep the line numbers visible
			if (codeViewContainer.classList.contains('wrap-enabled')) {
				if (lineInfoDiv) lineInfoDiv.style.display = 'none';
			} else {
				if (lineInfoDiv) lineInfoDiv.style.display = '';
			}
		});

		const searchInput = document.getElementById('searchInput');
		const prevMatchButton = document.getElementById('prevMatchButton');
		const nextMatchButton = document.getElementById('nextMatchButton');
		const searchResultCount = document.getElementById('searchResultCount');
		let matches = [];
		let currentMatchIndex = -1;

		function escapeHTML(str) {
			return str
				.replace(/&/g, "&amp;")
				.replace(/</g, "&lt;")
				.replace(/>/g, "&gt;")
				.replace(/"/g, "&quot;")
				.replace(/'/g, "&#039;");
		}

		const originalHTML = codeBlock.innerHTML; // cache only once

		const performSearch = () => {
			const searchTerm = searchInput.value.trim();
			if (!searchTerm) {
				codeBlock.innerHTML = originalHTML;
				hljs.highlightElement(codeBlock);
				searchResultCount.textContent = '';
				matches = [];
				currentMatchIndex = -1;
				return;
			}

			// Reset to original content and apply syntax highlighting first
			codeBlock.innerHTML = originalHTML;
			hljs.highlightElement(codeBlock);

			// Now search in the text content only, not HTML
			const textContent = codeBlock.textContent || '';
			const searchRegex = new RegExp(searchTerm.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi');
			
			// Find all text matches
			let match;
			const textMatches = [];
			while ((match = searchRegex.exec(textContent)) !== null) {
				textMatches.push({
					start: match.index,
					end: match.index + match[0].length,
					text: match[0]
				});
			}

			if (textMatches.length === 0) {
				searchResultCount.textContent = '0 results';
				matches = [];
				currentMatchIndex = -1;
				return;
			}

			// Use TreeWalker to find text nodes and wrap matches
			const walker = document.createTreeWalker(
				codeBlock,
				NodeFilter.SHOW_TEXT,
				null,
				false
			);

			const textNodes = [];
			let node;
			while (node = walker.nextNode()) {
				textNodes.push(node);
			}

			// Process matches from end to start to avoid offset issues
			textMatches.reverse().forEach((matchInfo, index) => {
				let currentOffset = 0;
				for (const textNode of textNodes) {
					const nodeLength = textNode.textContent.length;
					if (currentOffset + nodeLength > matchInfo.start) {
						const relativeStart = Math.max(0, matchInfo.start - currentOffset);
						const relativeEnd = Math.min(nodeLength, matchInfo.end - currentOffset);
						
						if (relativeStart < relativeEnd) {
							const range = document.createRange();
							range.setStart(textNode, relativeStart);
							range.setEnd(textNode, relativeEnd);
							
							const mark = document.createElement('mark');
							mark.setAttribute('data-idx', textMatches.length - 1 - index);
							
							try {
								range.surroundContents(mark);
							} catch (e) {
								// If surroundContents fails, extract and wrap
								const contents = range.extractContents();
								mark.appendChild(contents);
								range.insertNode(mark);
							}
						}
						break;
					}
					currentOffset += nodeLength;
				}
			});

			matches = Array.from(codeBlock.querySelectorAll('mark'));

			if (matches.length > 0) {
				currentMatchIndex = 0;
				updateActiveMatch();
			} else {
				currentMatchIndex = -1;
				searchResultCount.textContent = '0 results';
			}
		};



		const updateActiveMatch = () => {
			matches.forEach(m => m.classList.remove('active'));
			if (currentMatchIndex >= 0 && matches[currentMatchIndex]) {
				const match = matches[currentMatchIndex];
				match.classList.add('active');
				match.scrollIntoView({ block: 'center', behavior: 'smooth' });
				searchResultCount.textContent = `${currentMatchIndex + 1}/${matches.length} results`;
			}
		};


		const navigateSearch = (direction) => {
			if (!matches.length) return;
			currentMatchIndex = (currentMatchIndex + direction + matches.length) % matches.length;
			updateActiveMatch();
		};

		const searchForm = searchInput.closest('form');
		if (searchForm) {
			searchForm.addEventListener('submit', (e) => {
				e.preventDefault();
				e.stopPropagation();
				return false;
			});
		}
		searchInput.addEventListener('keydown', (e) => {
			if (e.key === 'Enter') {
				e.preventDefault();
				e.stopPropagation();
				performSearch();
				return false;
			}
			if (e.key === 'ArrowUp') {
				e.preventDefault();
				navigateSearch(-1);
			}
			if (e.key === 'ArrowDown') {
				e.preventDefault();
				navigateSearch(1);
			}
		});
		
		// Add event listeners (remove duplicates)
		if (searchButton) searchButton.addEventListener('click', performSearch);
		if (prevMatchButton) prevMatchButton.addEventListener('click', () => navigateSearch(-1));
		if (nextMatchButton) nextMatchButton.addEventListener('click', () => navigateSearch(1));
		
		if (searchInput) {
			searchInput.addEventListener('keyup', (e) => {
				if (e.key === 'Enter') performSearch();
			});
		}
	}
}

