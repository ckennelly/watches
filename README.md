watches - An n-way file tree differencer
(c) 2015 - Chris Kennelly (chris@ckennelly.com)

Overview
========

> A man with one watch always knows what time it is.  A man with two is never
> quite sure.

`watches` searches file trees and identifies points of difference.  Files are
compared using sha256, which makes false negatives (all files agreeing despite
having differences) unlikely but not impossible.
