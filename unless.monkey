let unless = macro(condition, consequence, alternative) {
    quote(
        if (unquote(condition)) {
            unquote(alternative);
        } else {
            unquote(consequence);
        }
        );
}
