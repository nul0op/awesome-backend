package model

const GH_API_URL = "https://api.github.com"
const AW_ROOT = "https://github.com/sindresorhus/awesome"

const RC_OK = 0x0                         // success
const RC_LINK_HAS_NO_INDEX_PAGE = 0x1     // the provided has no index page (currently only readme type is implemented)
const RC_LINK_IS_NOT_A_PROJECT_ROOT = 0x2 // link is towards a either a user landing page or a subfolder someware under the project root
const RC_LINK_IS_NOT_ON_GITHUB = 0x3
