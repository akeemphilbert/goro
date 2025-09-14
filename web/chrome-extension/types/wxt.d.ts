// WXT function declarations
declare function defineBackground(fn: () => void): any;
declare function defineContentScript(config: { matches: string[]; main: () => void }): any;
declare function defineUnlistedScript(fn: () => void): any;