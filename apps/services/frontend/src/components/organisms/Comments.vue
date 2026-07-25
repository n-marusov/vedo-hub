<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism/Comments — flat list of comment entries, no tabs, no card wrapper -->
<!-- Matches design/ui-kit.lib.pen orgComments (Organism/Comments) -->
<template>
  <div class="comments" role="region" aria-label="Comments">
    <CommentItem
      v-for="comment in comments"
      :key="comment.author + comment.timestamp"
      :author="comment.author"
      :handle="comment.handle"
      :timestamp="comment.timestamp"
      :action="comment.action"
      :text="comment.text"
    />
  </div>
</template>

<script setup lang="ts">
import CommentItem from "./CommentItem.vue";

interface CommentEntry {
	author: string;
	handle: string;
	avatarSrc?: string;
	timestamp: string;
	action: string;
	text: string;
}

withDefaults(
	defineProps<{
		comments?: CommentEntry[];
	}>(),
	{
		comments: () => [
			{
				author: "Nikolay Marusov",
				handle: "@nikomaru",
				timestamp: "2 hours ago",
				action:
					'commented on merge request !1 "Draft: Test 2" at Настоящее образование / Philosophy',
				text: "Проверка",
			},
			{
				author: "Anna Petrova",
				handle: "@anna",
				timestamp: "5 hours ago",
				action:
					'commented on merge request !2 "Fix: Validation rules" at VEDO / VEDO Core',
				text: "Need to update the SHACL constraints for the new release.",
			},
			{
				author: "Ivan Sidorov",
				handle: "@ivan",
				timestamp: "1 day ago",
				action:
					'commented on merge request !3 "Fix: Serialization bug" at VEDO / VEDO Core',
				text: "The issue was in the RDF/XML writer. Fixed in the latest commit.",
			},
		],
	},
);
</script>

<style scoped>
.comments {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 0;
}
</style>
