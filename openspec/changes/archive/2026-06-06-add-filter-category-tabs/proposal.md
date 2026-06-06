We currently have PresenterFilterPanel.tsx with both presenter and category 
filter. It consists of two main parts the actual filter options and an 
element at the bottom saying "Spots van" with the current count of selected 
presenters. Let's refer to this as the Filter Options and the Filter Status.

What I want is for only the Filter Status to be 
swipeable/scrollable tabs. We already use shadcn so we can probably use 
https://www.shadcn.io/examples/scrollable-tabs. The tabs that we can scroll 
between would be the presenter and category tabs. For the category tabs in 
the filter status, I would like to introduce the concept of the primary 
status and additional status. The primary status will be "SPOTS VAN" and 
"CATEGORIEËN", the additional status will be "X van Y Amsterdammers" and "X 
van Y categorieën". When a tab is active it should show both primary and 
addditional status. For the inactive tabs, it should only show the primary 
status, and the primary status of the inactive tab should be greyed out. 
Please use css transitions for both the greying out and showing or not 
showing the additional status.

You should have playwright-cli available, please interact with the 
filters/scroll as a user to verify your changes.