/*
 * ==========================================================================
 * UI state
 * ==========================================================================
 */

const UI_STORAGE_KEYS = {
    theme: "systemMonitor.theme",
    sections: "systemMonitor.sections"
};

const DEFAULT_SECTION_STATE = {
    system: true,

    cpu: true,
    cpuCores: false,

    memory: true,

    disk: true,
    diskIO: false,

    network: true,

    logs: false
};

let sectionState = loadSectionState();


/*
 * ==========================================================================
 * Initialization
 * ==========================================================================
 */

document.addEventListener("DOMContentLoaded", () => {
    initializeTheme();
    initializeCollapsibleSections();
    initializeBackToTop();
});


/*
 * ==========================================================================
 * Theme
 * ==========================================================================
 */

function initializeTheme() {
    const lightButton = document.getElementById("themeLight");
    const darkButton = document.getElementById("themeDark");

    if (lightButton) {
        lightButton.addEventListener("click", () => {
            setTheme("light", true);
        });
    }

    if (darkButton) {
        darkButton.addEventListener("click", () => {
            setTheme("dark", true);
        });
    }

    updateThemeButtons();

    /*
     * If the user has not explicitly selected a theme,
     * follow the operating system preference.
     */
    const mediaQuery = window.matchMedia(
        "(prefers-color-scheme: light)"
    );

    const handleSystemThemeChange = event => {
        /*
         * Once the user has selected Light/Dark manually,
         * system changes must no longer override that choice.
         */
        if (!localStorage.getItem(UI_STORAGE_KEYS.theme)) {
            applyTheme(event.matches ? "light" : "dark");
            updateThemeButtons();
        }
    };

    if (typeof mediaQuery.addEventListener === "function") {
        mediaQuery.addEventListener(
            "change",
            handleSystemThemeChange
        );
    } else if (typeof mediaQuery.addListener === "function") {
        mediaQuery.addListener(handleSystemThemeChange);
    }
}

function setTheme(theme, persist = true) {
    if (theme !== "light" && theme !== "dark") {
        return;
    }

    applyTheme(theme);

    if (persist) {
        try {
            localStorage.setItem(
                UI_STORAGE_KEYS.theme,
                theme
            );
        } catch (error) {
            console.warn(
                "Unable to persist theme preference:",
                error
            );
        }
    }

    updateThemeButtons();

    /*
     * IMPORTANT:
     *
     * Do not modify Chart.js objects here.
     *
     * Chart.js keeps internal Proxy objects and modifying its
     * configuration while it is updating/resizing can cause:
     *
     *     InternalError: too much recursion
     *
     * The page theme itself is handled entirely by CSS.
     */
}

function applyTheme(theme) {
    document.documentElement.dataset.theme = theme;
}

function updateThemeButtons() {
    const currentTheme =
        document.documentElement.dataset.theme || "dark";

    const lightButton = document.getElementById("themeLight");
    const darkButton = document.getElementById("themeDark");

    if (lightButton) {
        lightButton.setAttribute(
            "aria-pressed",
            String(currentTheme === "light")
        );
    }

    if (darkButton) {
        darkButton.setAttribute(
            "aria-pressed",
            String(currentTheme === "dark")
        );
    }
}


/*
 * ==========================================================================
 * Collapsible sections
 * ==========================================================================
 */

function initializeCollapsibleSections() {
    const buttons = document.querySelectorAll(
        "[data-collapse-target]"
    );

    buttons.forEach(button => {
        const section = button.closest("[data-section]");

        if (!section) {
            return;
        }

        const sectionName = section.dataset.section;
        const targetId = button.dataset.collapseTarget;
        const target = document.getElementById(targetId);

        if (!target || !sectionName) {
            return;
        }

        /*
         * localStorage has priority over the HTML default.
         */
        const expanded =
            Object.prototype.hasOwnProperty.call(
                sectionState,
                sectionName
            )
                ? Boolean(sectionState[sectionName])
                : button.getAttribute("aria-expanded") !== "false";

        sectionState[sectionName] = expanded;

        setCollapsibleState(
            button,
            target,
            expanded
        );

        button.addEventListener("click", () => {
            const nextState =
                button.getAttribute("aria-expanded") !== "true";

            sectionState[sectionName] = nextState;

            saveSectionState();

            setCollapsibleState(
                button,
                target,
                nextState
            );
        });
    });

    saveSectionState();
}

function setCollapsibleState(button, target, expanded) {
    button.setAttribute(
        "aria-expanded",
        String(expanded)
    );

    target.classList.toggle(
        "is-collapsed",
        !expanded
    );
}

function loadSectionState() {
    try {
        const raw = localStorage.getItem(
            UI_STORAGE_KEYS.sections
        );

        if (!raw) {
            return { ...DEFAULT_SECTION_STATE };
        }

        const saved = JSON.parse(raw);

        if (!saved || typeof saved !== "object") {
            return { ...DEFAULT_SECTION_STATE };
        }

        const result = {
            ...DEFAULT_SECTION_STATE
        };

        for (const key of Object.keys(DEFAULT_SECTION_STATE)) {
            if (typeof saved[key] === "boolean") {
                result[key] = saved[key];
            }
        }

        return result;
    } catch (error) {
        console.warn(
            "Unable to load section state:",
            error
        );

        return { ...DEFAULT_SECTION_STATE };
    }
}

function saveSectionState() {
    try {
        localStorage.setItem(
            UI_STORAGE_KEYS.sections,
            JSON.stringify(sectionState)
        );
    } catch (error) {
        console.warn(
            "Unable to persist section state:",
            error
        );
    }
}


/*
 * ==========================================================================
 * Russian date/time formatting
 * ==========================================================================
 *
 * Required format:
 *
 *     DD.MM.YYYY HH:mm:ss
 *
 * We construct the string manually instead of relying on
 * toLocaleString(), so the result is identical in every browser.
 */

function formatTime(timestamp) {
    if (!timestamp) {
        return "—";
    }

    const date = new Date(timestamp);

    if (Number.isNaN(date.getTime())) {
        return "—";
    }

    const hours = String(date.getHours()).padStart(2, "0");
    const minutes = String(date.getMinutes()).padStart(2, "0");
    const seconds = String(date.getSeconds()).padStart(2, "0");

    return `${hours}:${minutes}:${seconds}`;
}

function formatDateTime(timestamp) {
    if (!timestamp) {
        return "—";
    }

    const date = new Date(timestamp);

    if (Number.isNaN(date.getTime())) {
        return "—";
    }

    const day = String(date.getDate()).padStart(2, "0");
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const year = String(date.getFullYear());

    const hours = String(date.getHours()).padStart(2, "0");
    const minutes = String(date.getMinutes()).padStart(2, "0");
    const seconds = String(date.getSeconds()).padStart(2, "0");

    return (
        `${day}.${month}.${year} ` +
        `${hours}:${minutes}:${seconds}`
    );
}

/*
 * ==========================================================================
 * Back to top
 * ==========================================================================
 */

function initializeBackToTop() {
    const button = document.getElementById("backToTop");

    if (!button) {
        return;
    }

    const scrollThreshold = 400;

    const updateVisibility = () => {
        button.classList.toggle(
            "visible",
            window.scrollY > scrollThreshold
        );
    };

    window.addEventListener(
        "scroll",
        updateVisibility,
        { passive: true }
    );

    button.addEventListener("click", () => {
        window.scrollTo({
            top: 0,
            behavior: "smooth"
        });
    });

    updateVisibility();
}