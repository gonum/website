---
# Display name
title: Gonum Numerical Packages

# Username (this should match the folder name)
authors:
- admin

# Is this the primary user of the site?
superuser: true

# Role/position
role: Consistent, composable, and comprehensible scientific code

# Organizations/Affiliations
organizations:
- name: ""
  url: ""

# Short bio (displayed in user profile at end of posts)
bio: ""

# Social/Academic Networking
# For available icons, see: https://sourcethemes.com/academic/docs/page-builder/#icons
#   For an email link, use "fas" icon pack, "envelope" icon, and a link in the
#   form "mailto:your-email@example.com" or "#contact" for contact widget.
social:
- icon: envelope
  icon_pack: fas
  link: "mailto:gonum-dev@googlegroups.com"  # For a form, use  '#contact'
#- icon: twitter
#  icon_pack: fab
#  link: https://twitter.com/gonum
#- icon: google-scholar
#  icon_pack: ai
#  link: ""
- icon: github
  icon_pack: fab
  link: https://github.com/gonum
- icon: book
  icon_pack: fa
  link: https://pkg.go.dev/mod/gonum.org/v1/gonum
# Link to a PDF of your resume/CV from the About widget.
# To enable, copy your resume/CV to `static/files/cv.pdf` and uncomment the lines below.
# - icon: cv
#   icon_pack: ai
#   link: files/cv.pdf

# Enter email to display Gravatar (if Gravatar enabled in Config)
email: ""

# Organizational groups that you belong to (for People widget)
#   Set this to `[]` or comment out if you are not using People widget.
user_groups:
- Researchers
- Visitors
---

Gonum is a set of packages designed to make writing numerical and scientific
algorithms productive, performant, and scalable.

Gonum contains libraries for [matrices and linear algebra](https://pkg.go.dev/gonum.org/v1/gonum/mat);
[statistics](https://pkg.go.dev/gonum.org/v1/gonum/stat), 
[probability](https://pkg.go.dev/gonum.org/v1/gonum/stat/distuv) 
[distributions](https://pkg.go.dev/gonum.org/v1/gonum/stat/distmv), 
and [sampling](https://pkg.go.dev/gonum.org/v1/gonum/stat/sampleuv); tools for
[function differentiation](https://pkg.go.dev/gonum.org/v1/gonum/diff/fd), 
[integration](https://pkg.go.dev/gonum.org/v1/gonum/integrate/quad),
and [optimization](https://pkg.go.dev/gonum.org/v1/gonum/optimize);
[network](https://pkg.go.dev/gonum.org/v1/gonum/graph) creation and analysis; and more.

We encourage you to [get started](post/intro_to_gonum) with Go and Gonum if

* You are tired of sluggish performance, and fighting C and vectorization.
* You are struggling with managing programs as they grow larger.
* You struggle to re-use -- even the code you tried to make reusable.
* You would like easy access to parallel computing.
* You want code to be fully transparent, and want the ability to read the source code you use.
* You'd like a compiler to catch mistakes early, but hate fighting linker and unintelligible compile errors.

