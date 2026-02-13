// deeply-nested.ts — Stress test: deep object nesting with optional chaining.

export interface DeepConfig {
  level1: {
    level2: {
      level3: {
        level4: {
          level5: {
            level6: {
              level7: {
                level8: {
                  level9: {
                    level10: {
                      value: string;
                    };
                  };
                };
              };
            };
          };
        };
      };
    };
  };
}

export function getDeepValue(config: DeepConfig): string | undefined {
  return config?.level1?.level2?.level3?.level4?.level5?.level6?.level7?.level8?.level9?.level10?.value;
}

export function setDeepValue(config: DeepConfig, value: string): void {
  config.level1.level2.level3.level4.level5.level6.level7.level8.level9.level10.value = value;
}
