import classes from "./style.module.css";

const $app = document.querySelector<HTMLDivElement>("#app")!;

function parseWords(text: string) {
  const s = text.split(" ");
  return s;
}

function wrapWordsInSpan(text: string, delayOffset = 0) {
  const s = parseWords(text);
  let newText = "";
  let animationDelay = 0;
  s.forEach((word) => {
    newText += `<span style="animation-delay: ${delayOffset + animationDelay * 200}ms, ${delayOffset + 1000 + animationDelay * 200}ms">${word}&nbsp;</span>`;
    animationDelay += 1;
  });
  return newText;
}

const templateHtml = `
<div class="${classes.wrapper}">
  <div class="${classes.mainMessage}">
    <div class="${classes.letterFade}">
      <div class="${classes.morningMessage}">
        ${wrapWordsInSpan("Good Morning")}
      </div>
      <div class="${classes.intentMessage}">
        ${wrapWordsInSpan("Let's set your intent for the day.", 2000)}
      </div>
    </div>
    <button id="begin-btn" class="${classes.buttonCtl}">Begin</button>
  </div>
</div>
`;

interface View {
  getDom(): Element;
  loadView(): void;
  getData(): string;
}

class GoodMorning implements View {
  private _$el: Element;
  private _$beginBtn: HTMLButtonElement | null = null;

  constructor() {
    const template = document.createElement("template");
    template.innerHTML = templateHtml;
    const fragment = template.content.cloneNode(true) as DocumentFragment;
    this._$el = fragment.firstElementChild!;
  }

  getDom() {
    return this._$el;
  }

  loadView() {
    this._$beginBtn = this._$el.querySelector("#begin-btn");

    this._addEvents();
  }

  getData(): string {
    return "";
  }

  _addEvents() {
    this._$beginBtn?.addEventListener("click", this._handleBegin);
  }

  _removeEvents() {
    this._$beginBtn?.removeEventListener("click", this._handleBegin);
  }

  _handleBegin = (_event: Event) => {
    pageState.navigateToNextPage();
  };
}

const templateHtml1 = `
<div class="${classes.wrapper}">
  <div class="${classes.mainMessage}">
    <div class="${classes.queryWrapper}">
   <div class="${classes.query}">
    ${wrapWordsInSpan("Yesterday... ")}
    </div>
   <div class="${classes.query}">
    ${wrapWordsInSpan("what did you intend to do?", 1000)}
    </div>
    </div>
    <div class="${classes.entryWrapper}">
      <div contenteditable="plaintext-only" id="yesterday-intent" class="${classes.textarea}"></div>
      <button id="continue-button" disabled class="${classes.continueBtn}" style="opacity: 0">Continue</button>
    </div>
  </div>
</div>
`;

class YesterdayIntent implements View {
  private _$el: Element;
  private _$yesterdayIntent: HTMLElement | null = null;
  private _$continueButton: HTMLButtonElement | null = null;
  private _intent = "";

  constructor() {
    let template = document.createElement("template");
    template.innerHTML = templateHtml1;
    const fragment = template.content.cloneNode(true) as DocumentFragment;
    this._$el = fragment.firstElementChild!;
  }

  getDom() {
    return this._$el;
  }

  loadView() {
    this._$yesterdayIntent = this._$el.querySelector("#yesterday-intent");
    this._$continueButton = this._$el.querySelector("#continue-button");

    this._$yesterdayIntent?.focus();

    this._addEvents();
  }

  getData(): string {
    return `Yesterdays intent: ${this._intent}`;
  }

  _addEvents() {
    this._$yesterdayIntent?.addEventListener("input", this._handleFieldChange);
    this._$continueButton?.addEventListener("click", this._handleContinue);
  }

  _removeEvents() {
    this._$yesterdayIntent?.removeEventListener(
      "input",
      this._handleFieldChange,
    );
    this._$continueButton?.removeEventListener("click", this._handleContinue);
  }

  _handleContinue = (_event: Event) => {
    console.log("continue");
    pageState.navigateToNextPage();
  };

  _handleFieldChange = (event: Event) => {
    const target = event.target as HTMLElement;
    console.log(target.textContent);
    this._intent = target.textContent || "";
    this._updateView();
  };

  _updateView() {
    if (this._intent.length) {
      this._$continueButton!.style.opacity = "1";
      this._$continueButton!.disabled = false;
    } else {
      this._$continueButton!.style.opacity = "0";
      this._$continueButton!.disabled = true;
    }
  }
}

const templateHtml2 = `
<div class="${classes.wrapper}">
  <div class="${classes.mainMessage}">
    <div class="${classes.queryWrapper}">
   <div class="${classes.query}">
    ${wrapWordsInSpan("What did you complete?")}
    </div>
    </div>
    <div class="${classes.entryWrapper}" style="--entry-delay: 1.5s; --entry-fade-delay: 2s">
      <div contenteditable="plaintext-only" id="yesterday-intent" class="${classes.textarea}"></div>
      <button id="continue-button" disabled class="${classes.continueBtn}" style="opacity: 0">Continue</button>
    </div>
  </div>
</div>
`;

class YesterdayHappened implements View {
  private _$el: Element;
  private _$yesterdayIntent: HTMLElement | null = null;
  private _$continueButton: HTMLButtonElement | null = null;
  private _intent = "";

  constructor() {
    let template = document.createElement("template");
    template.innerHTML = templateHtml2;
    const fragment = template.content.cloneNode(true) as DocumentFragment;
    this._$el = fragment.firstElementChild!;
  }

  getDom() {
    return this._$el;
  }

  loadView() {
    this._$yesterdayIntent = this._$el.querySelector("#yesterday-intent");
    this._$continueButton = this._$el.querySelector("#continue-button");

    this._$yesterdayIntent?.focus();

    this._addEvents();
  }

  getData(): string {
    return `Yesterday's completed: ${this._intent}`;
  }

  _addEvents() {
    this._$yesterdayIntent?.addEventListener("input", this._handleFieldChange);
    this._$continueButton?.addEventListener("click", this._handleContinue);
  }

  _removeEvents() {
    this._$yesterdayIntent?.removeEventListener(
      "input",
      this._handleFieldChange,
    );
    this._$continueButton?.removeEventListener("click", this._handleContinue);
  }

  _handleContinue = (_event: Event) => {
    console.log("continue");
    pageState.navigateToNextPage();
  };

  _handleFieldChange = (event: Event) => {
    const target = event.target as HTMLElement;
    console.log(target.textContent);
    this._intent = target.textContent || "";
    this._updateView();
  };

  _updateView() {
    if (this._intent.length) {
      this._$continueButton!.style.opacity = "1";
      this._$continueButton!.disabled = false;
    } else {
      this._$continueButton!.style.opacity = "0";
      this._$continueButton!.disabled = true;
    }
  }
}

const templateHtml3 = `
<div class="${classes.wrapper}">
  <div class="${classes.mainMessage}">
    <div class="${classes.queryWrapper}">
   <div class="${classes.query}">
    ${wrapWordsInSpan("Was there friction?")}
    </div>
    </div>
    <div class="${classes.entryWrapper}" style="--entry-delay: 1.5s; --entry-fade-delay: 2s">
      <div contenteditable="plaintext-only" id="yesterday-intent" class="${classes.textarea}"></div>
      <button id="continue-button" disabled class="${classes.continueBtn}" style="opacity: 0">Continue</button>
    </div>
  </div>
</div>
`;

class YesterdayFriction implements View {
  private _$el: Element;
  private _$yesterdayIntent: HTMLElement | null = null;
  private _$continueButton: HTMLButtonElement | null = null;
  private _intent = "";

  constructor() {
    let template = document.createElement("template");
    template.innerHTML = templateHtml3;
    const fragment = template.content.cloneNode(true) as DocumentFragment;
    this._$el = fragment.firstElementChild!;
  }

  getDom() {
    return this._$el;
  }

  loadView() {
    this._$yesterdayIntent = this._$el.querySelector("#yesterday-intent");
    this._$continueButton = this._$el.querySelector("#continue-button");

    this._$yesterdayIntent?.focus();

    this._addEvents();
  }

  getData(): string {
    return `Yesterday's friction: ${this._intent}`;
  }

  _addEvents() {
    this._$yesterdayIntent?.addEventListener("input", this._handleFieldChange);
    this._$continueButton?.addEventListener("click", this._handleContinue);
  }

  _removeEvents() {
    this._$yesterdayIntent?.removeEventListener(
      "input",
      this._handleFieldChange,
    );
    this._$continueButton?.removeEventListener("click", this._handleContinue);
  }

  _handleContinue = (_event: Event) => {
    console.log("continue");
    pageState.navigateToNextPage();
  };

  _handleFieldChange = (event: Event) => {
    const target = event.target as HTMLElement;
    console.log(target.textContent);
    this._intent = target.textContent || "";
    this._updateView();
  };

  _updateView() {
    if (this._intent.length) {
      this._$continueButton!.style.opacity = "1";
      this._$continueButton!.disabled = false;
    } else {
      this._$continueButton!.style.opacity = "0";
      this._$continueButton!.disabled = true;
    }
  }
}

const templateHtml4 = `
<div class="${classes.wrapper}">
  <div class="${classes.mainMessage}">
    <div class="${classes.queryWrapper}">
     <div class="${classes.query}">
    ${wrapWordsInSpan("Today...")}
    </div>
   <div class="${classes.query}">
    ${wrapWordsInSpan("What is your intent?", 1000)}
    </div>
    </div>
    <div class="${classes.entryWrapper}">
      <div contenteditable="plaintext-only" id="yesterday-intent" class="${classes.textarea}"></div>
      <button id="continue-button" disabled class="${classes.continueBtn}" style="opacity: 0">Continue</button>
    </div>
  </div>
</div>
`;

class TodayIntent implements View {
  private _$el: Element;
  private _$yesterdayIntent: HTMLElement | null = null;
  private _$continueButton: HTMLButtonElement | null = null;
  private _intent = "";

  constructor() {
    let template = document.createElement("template");
    template.innerHTML = templateHtml4;
    const fragment = template.content.cloneNode(true) as DocumentFragment;
    this._$el = fragment.firstElementChild!;
  }

  getDom() {
    return this._$el;
  }

  loadView() {
    this._$yesterdayIntent = this._$el.querySelector("#yesterday-intent");
    this._$continueButton = this._$el.querySelector("#continue-button");

    this._$yesterdayIntent?.focus();

    this._addEvents();
  }

  getData(): string {
    return `Today's intent: ${this._intent}`;
  }

  _addEvents() {
    this._$yesterdayIntent?.addEventListener("input", this._handleFieldChange);
    this._$continueButton?.addEventListener("click", this._handleContinue);
  }

  _removeEvents() {
    this._$yesterdayIntent?.removeEventListener(
      "input",
      this._handleFieldChange,
    );
    this._$continueButton?.removeEventListener("click", this._handleContinue);
  }

  _handleContinue = (_event: Event) => {
    console.log("continue");
    pageState.navigateToNextPage();
  };

  _handleFieldChange = (event: Event) => {
    const target = event.target as HTMLElement;
    console.log(target.textContent);
    this._intent = target.textContent || "";
    this._updateView();
  };

  _updateView() {
    if (this._intent.length) {
      this._$continueButton!.style.opacity = "1";
      this._$continueButton!.disabled = false;
    } else {
      this._$continueButton!.style.opacity = "0";
      this._$continueButton!.disabled = true;
    }
  }
}

const templateHtml5 = `
<div class="${classes.wrapper}">
  <div class="${classes.mainMessage}">
    <div class="${classes.queryWrapper}">
     <div class="${classes.query}">
    ${wrapWordsInSpan("Are you blocked, need to clarify, or have concerns?")}
    </div>
    </div>
    <div class="${classes.entryWrapper}" style="--entry-delay: 2s; --entry-fade-delay: 2.5s">
      <div contenteditable="plaintext-only" id="yesterday-intent" class="${classes.textarea}"></div>
      <button id="continue-button" disabled class="${classes.continueBtn}" style="opacity: 0">Continue</button>
    </div>
  </div>
</div>
`;

class TodayQuestions implements View {
  private _$el: Element;
  private _$yesterdayIntent: HTMLElement | null = null;
  private _$continueButton: HTMLButtonElement | null = null;
  private _intent = "";

  constructor() {
    let template = document.createElement("template");
    template.innerHTML = templateHtml5;
    const fragment = template.content.cloneNode(true) as DocumentFragment;
    this._$el = fragment.firstElementChild!;
  }

  getDom() {
    return this._$el;
  }

  loadView() {
    this._$yesterdayIntent = this._$el.querySelector("#yesterday-intent");
    this._$continueButton = this._$el.querySelector("#continue-button");

    this._$yesterdayIntent?.focus();

    this._addEvents();
  }

  getData(): string {
    return `Today's questions or concerns: ${this._intent}`;
  }

  _addEvents() {
    this._$yesterdayIntent?.addEventListener("input", this._handleFieldChange);
    this._$continueButton?.addEventListener("click", this._handleContinue);
  }

  _removeEvents() {
    this._$yesterdayIntent?.removeEventListener(
      "input",
      this._handleFieldChange,
    );
    this._$continueButton?.removeEventListener("click", this._handleContinue);
  }

  _handleContinue = (_event: Event) => {
    console.log("continue");
    pageState.navigateToNextPage();
  };

  _handleFieldChange = (event: Event) => {
    const target = event.target as HTMLElement;
    console.log(target.textContent);
    this._intent = target.textContent || "";
    this._updateView();
  };

  _updateView() {
    if (this._intent.length) {
      this._$continueButton!.style.opacity = "1";
      this._$continueButton!.disabled = false;
    } else {
      this._$continueButton!.style.opacity = "0";
      this._$continueButton!.disabled = true;
    }
  }
}

const templateHtml6 = `
<div class="${classes.wrapper}">
  <div class="${classes.mainMessage}">
    <div class="${classes.queryWrapper}">
     <div class="${classes.query}">
    ${wrapWordsInSpan("Looking forward...")}
    </div>
     <div class="${classes.query}">
    ${wrapWordsInSpan("What are you goals and hopes?", 1000)}
    </div>
    </div>
    <div class="${classes.entryWrapper}">
      <div contenteditable="plaintext-only" id="yesterday-intent" class="${classes.textarea}"></div>
      <button id="continue-button" disabled class="${classes.continueBtn}" style="opacity: 0">Continue</button>
    </div>
  </div>
</div>
`;

class Goal implements View {
  private _$el: Element;
  private _$yesterdayIntent: HTMLElement | null = null;
  private _$continueButton: HTMLButtonElement | null = null;
  private _intent = "";

  constructor() {
    let template = document.createElement("template");
    template.innerHTML = templateHtml6;
    const fragment = template.content.cloneNode(true) as DocumentFragment;
    this._$el = fragment.firstElementChild!;
  }

  getDom() {
    return this._$el;
  }

  loadView() {
    this._$yesterdayIntent = this._$el.querySelector("#yesterday-intent");
    this._$continueButton = this._$el.querySelector("#continue-button");

    this._$yesterdayIntent?.focus();

    this._addEvents();
  }

  getData(): string {
    return `Big Piciture Goal or thoughts: ${this._intent}`;
  }

  _addEvents() {
    this._$yesterdayIntent?.addEventListener("input", this._handleFieldChange);
    this._$continueButton?.addEventListener("click", this._handleContinue);
  }

  _removeEvents() {
    this._$yesterdayIntent?.removeEventListener(
      "input",
      this._handleFieldChange,
    );
    this._$continueButton?.removeEventListener("click", this._handleContinue);
  }

  _handleContinue = (_event: Event) => {
    console.log("continue");
    pageState.navigateToNextPage();
  };

  _handleFieldChange = (event: Event) => {
    const target = event.target as HTMLElement;
    console.log(target.textContent);
    this._intent = target.textContent || "";
    this._updateView();
  };

  _updateView() {
    if (this._intent.length) {
      this._$continueButton!.style.opacity = "1";
      this._$continueButton!.disabled = false;
    } else {
      this._$continueButton!.style.opacity = "0";
      this._$continueButton!.disabled = true;
    }
  }
}

const backTemplate = `
  <div class="${classes.backButtonWrapper}">
    <button id="back-button" class="${classes.backButton} ${classes.backButtonHidden}">Back</button>
  </div>
`;

class BackButton {
  private _$el: Element;
  private _$btn: HTMLButtonElement | null = null;

  constructor() {
    let template = document.createElement("template");
    template.innerHTML = backTemplate;
    const fragment = template.content.cloneNode(true) as DocumentFragment;
    this._$el = fragment.firstElementChild!;
  }

  getDom() {
    return this._$el;
  }

  loadView() {
    this._$btn = this._$el.querySelector("#back-button");
    this._addEvents();
    // Button starts hidden via CSS class, no need to call updateVisibility here
  }

  updateVisibility(show: boolean) {
    if (this._$btn) {
      if (show) {
        this._$btn.classList.remove(classes.backButtonHidden);
        this._$btn.classList.add(classes.backButtonVisible);
      } else {
        this._$btn.classList.remove(classes.backButtonVisible);
        this._$btn.classList.add(classes.backButtonHidden);
      }
    }
  }

  _addEvents() {
    this._$btn?.addEventListener("click", this._handlePress);
  }

  _removeEvents() {
    this._$btn?.removeEventListener("click", this._handlePress);
  }

  _handlePress = (_event: Event) => {
    console.log("back button pressed");
    history.back();
  };
}

const templateHtml7 = `
<div class="${classes.wrapper}">
  <div class="${classes.mainMessage}">
    <div class="${classes.queryWrapper}">
     <div class="${classes.query}">
    ${wrapWordsInSpan("Thanks for being you.")}
    </div>
     <div class="${classes.query}">
    ${wrapWordsInSpan("Here's your summary...", 2000)}
    </div>
    </div>
    <div class="${classes.summaryWrapper}">
      <div class="${classes.days}">
        <div class="${classes.day}">
          <em class="${classes.dayEmphasis}">Yesterday</em>
          <span class="">
            you finished the Metrics Page but were slowed by meetings and conflicting pressures
          </span>
        </div>
        <div class="${classes.day}">
          <em class="${classes.dayEmphasis}">Today</em>
          <span>
            you’re focused on the Run Details Page, resolving versioning clarity, and doing SRE research outreach, while watching out for design-by-committee and scope creep.
          </span>
        </div>
        <div class="${classes.day}">
          <em class="${classes.dayEmphasis}">Looking forward</em>
          <span>
            you want calmer, iterative rhythms with better cross-functional clarity and balance between design detail and speed.
          </span>
        </div>
        <div class="${classes.daysFooter}">
          <button>Edit Answers</button>
        </div>
      </div>
      <div class="${classes.summaryFooter}">
        <button>Continue</button>
      </div>
    </div>
  </div>
</div>
`;

class Summary implements View {
  private _$el: Element;
  private _$yesterdayIntent: HTMLElement | null = null;
  private _$continueButton: HTMLButtonElement | null = null;

  constructor() {
    let template = document.createElement("template");
    template.innerHTML = templateHtml7;
    const fragment = template.content.cloneNode(true) as DocumentFragment;
    this._$el = fragment.firstElementChild!;
  }

  getDom() {
    return this._$el;
  }

  loadView() {
    // this._$yesterdayIntent = this._$el.querySelector("#yesterday-intent");
    // this._$continueButton = this._$el.querySelector("#continue-button");

    // this._$yesterdayIntent?.focus();

    this._addEvents();
  }

  _addEvents() {
    // this._$yesterdayIntent?.addEventListener("input", this._handleFieldChange);
    // this._$continueButton?.addEventListener("click", this._handleContinue);
  }

  _removeEvents() {
    // this._$yesterdayIntent?.removeEventListener(
    //   "input",
    //   this._handleFieldChange,
    // );
    // this._$continueButton?.removeEventListener("click", this._handleContinue);
  }

  _handleContinue = (_event: Event) => {
    pageState.navigateToNextPage();
  };

  _handleFieldChange = (event: Event) => {
    const target = event.target as HTMLElement;
    this._updateView();
  };

  _updateView() {}
}

class PageState {
  private _currentPage = 0;
  private _totalViewsAdded = 0;
  private _currentView: View;
  private _views: View[] = [];
  private _backButton: BackButton;

  constructor() {
    // this._views.push(new Summary());
    this._views.push(new GoodMorning());
    this._views.push(new YesterdayIntent());
    this._views.push(new YesterdayHappened());
    this._views.push(new YesterdayFriction());
    this._views.push(new TodayIntent());
    this._views.push(new TodayQuestions());
    this._views.push(new Goal());

    this._views.forEach((view) => {
      view.loadView();
    });

    const goodMorningView = this._views[0];
    $app.appendChild(goodMorningView.getDom());
    this._totalViewsAdded = this._totalViewsAdded + 1;
    this._currentView = goodMorningView;

    this._backButton = new BackButton();
    $app.appendChild(this._backButton.getDom());
    this._backButton.loadView();
    window.addEventListener("popstate", this._handlePopState);

    // Set initial history state
    history.replaceState({ page: 0 }, "");
  }

  _handlePopState = (event: PopStateEvent) => {
    console.log(event);
    if (event.state && typeof event.state.page === "number") {
      const targetPage = event.state.page;
      if (targetPage < this._currentPage) {
        // Going back
        this._navigateToPage(targetPage);
      } else if (targetPage > this._currentPage) {
        // Going forward
        this._navigateToPage(targetPage);
      }
    } else if (!event.state) {
      // No state means we're at the initial page
      this._navigateToPage(0);
    }
  };

  navigateToNextPage() {
    if (this._currentPage >= this._views.length - 1) return;

    const nextPage = this._currentPage + 1;
    this._navigateToPage(nextPage);
    history.pushState({ page: nextPage }, "");
    this._updateBackButtonVisibility();
  }

  _viewHasBeenAppended(_view: View) {
    return this._currentPage < this._totalViewsAdded;
  }

  navigateToPreviousPage() {
    if (this._currentPage < 1) return;

    const prevPage = this._currentPage - 1;
    this._navigateToPage(prevPage);
    history.pushState({ page: prevPage }, "");
    this._updateBackButtonVisibility();
  }

  _navigateToPage(targetPage: number) {
    this._views.forEach((view) => {
      console.log(view.getData());
    });
    if (
      targetPage < 0 ||
      targetPage >= this._views.length ||
      targetPage === this._currentPage
    )
      return;

    const currentView = this._currentView;
    const isGoingForward = targetPage > this._currentPage;

    // Hide current view with appropriate animation
    if (isGoingForward) {
      currentView.getDom().classList.add(`${classes.wrapperHide}`);
    } else {
      currentView.getDom().classList.add(`${classes.wrapperHideForward}`);
    }

    // Update current page and view
    this._currentPage = targetPage;
    this._currentView = this._views[this._currentPage];

    // Show target view
    if (isGoingForward) {
      this._currentView
        .getDom()
        .classList.remove(`${classes.wrapperHideForward}`);
    } else {
      this._currentView.getDom().classList.remove(`${classes.wrapperHide}`);
    }

    // Add view to DOM if it hasn't been added yet
    if (!this._viewHasBeenAppended(this._currentView)) {
      $app.appendChild(this._currentView.getDom());
      this._totalViewsAdded = this._totalViewsAdded + 1;
    }

    this._updateBackButtonVisibility();
  }

  private _updateBackButtonVisibility() {
    this._backButton.updateVisibility(this._currentPage > 0);
  }
}

const pageState = new PageState();
