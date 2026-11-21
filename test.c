#include <stdlib.h>

int num = 42;
char *str;

int main() {
  num = 67;
  str = (char *) malloc(4 * sizeof(char));
  str[0] = 't';
  str[1] = 'e';
  str[2] = 's';
  str[3] = 't';
  str = new;
  free(str);
}
